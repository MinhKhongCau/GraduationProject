package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"payment-service/pkg/httpclient"
)

// restBookingClient là implementation REST của BookingServiceClient.
// Gọi sang /internal/appointments/:id/confirm|fail của Booking Service
// thông qua Internal JWT Auth (M2M).
type restBookingClient struct {
	baseURL    string // http://booking-service:8083
	httpClient *httpclient.Client
}

// webhookPayload là body gửi sang Booking Service.
// Giữ cùng cấu trúc với WebhookRequest hiện có của booking-service.
type webhookPayload struct {
	AppointmentID string `json:"appointment_id"`
	Status        string `json:"status"` // "SUCCESS" hoặc "FAILED"
}

// NewRestBookingClient tạo REST implementation.
//
//	tokenProvider: thường là *internal_auth.TokenManager đã được khởi tạo ở main.go.
//	baseURL: URL nội bộ của booking-service, e.g. "http://booking-service:8083".
func NewRestBookingClient(baseURL string, tokenProvider httpclient.TokenProvider) BookingServiceClient {
	return &restBookingClient{
		baseURL: baseURL,
		httpClient: httpclient.New(httpclient.Options{
			MaxRetries:    3,
			TokenProvider: tokenProvider,
		}),
	}
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    Appointment `json:"data"`
}

// GetAppointment gọi internal API để lấy thông tin chi tiết của appointment.
func (c *restBookingClient) GetAppointment(ctx context.Context, appointmentID string) (*Appointment, error) {
	url := fmt.Sprintf("%s/internal/appointments/%s", c.baseURL, appointmentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("booking_client: create get request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("booking_client: get appointment %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("booking_client: appointment not found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("booking_client: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("booking_client: decode response: %w", err)
	}

	return &apiResp.Data, nil
}

// ConfirmAppointment gọi endpoint internal của Booking Service để xác nhận lịch hẹn.
func (c *restBookingClient) ConfirmAppointment(ctx context.Context, appointmentID string) error {
	return c.callWebhook(ctx, appointmentID, "SUCCESS")
}

// FailAppointment gọi endpoint internal của Booking Service để hủy lịch hẹn.
func (c *restBookingClient) FailAppointment(ctx context.Context, appointmentID string) error {
	return c.callWebhook(ctx, appointmentID, "FAILED")
}

// callWebhook là phương thức dùng chung gọi POST /internal/appointments/:id/webhook
// trên Booking Service thông qua internal auth.
//
// Gọi endpoint /internal thay vì /public vì:
//   - /public không yêu cầu xác thực → bất kỳ ai cũng gọi được (kể cả giả mạo).
//   - /internal được bảo vệ bởi M2M JWT middleware — chỉ services nội bộ được phép.
func (c *restBookingClient) callWebhook(ctx context.Context, appointmentID, status string) error {
	url := fmt.Sprintf("%s/internal/appointments/%s/webhook", c.baseURL, appointmentID)

	payload := webhookPayload{
		AppointmentID: appointmentID,
		Status:        status,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("booking_client: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("booking_client: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Authorization header sẽ được inject bởi httpclient.Client thông qua TokenProvider

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("booking_client: call %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("booking_client: unexpected status %d from booking-service: %s", resp.StatusCode, string(respBody))
	}

	log.Printf("✅ [booking_client] appointment %s → %s confirmed by booking-service", appointmentID, status)
	return nil
}
