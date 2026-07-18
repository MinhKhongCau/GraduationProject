package bookingrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	apppayment "payment-service/internal/payment/application"
	"payment-service/pkg/httpclient"

	"github.com/google/uuid"
)

type restBookingClient struct {
	baseURL    string
	httpClient *httpclient.Client
}

type webhookPayload struct {
	AppointmentID string `json:"appointment_id"`
	Status        string `json:"status"`
}

type paymentEligibilityRequest struct {
	PayerID string `json:"payer_id"`
}

type paymentEligibilityResponse struct {
	AppointmentID string `json:"appointment_id"`
	ExpertID      string `json:"expert_id"`
	AmountVND     int64  `json:"amount_vnd"`
	ExpiresAt     int64  `json:"expires_at"`
}

type paymentEligibilityAPIResponse struct {
	Success bool                       `json:"success"`
	Message string                     `json:"message"`
	Data    paymentEligibilityResponse `json:"data"`
	Error   string                     `json:"error"`
}

func NewRestBookingClient(baseURL string, tokenProvider httpclient.TokenProvider) apppayment.BookingServiceClient {
	return &restBookingClient{
		baseURL: baseURL,
		httpClient: httpclient.New(httpclient.Options{
			MaxRetries:    3,
			TokenProvider: tokenProvider,
		}),
	}
}

func (c *restBookingClient) GetPaymentEligibility(ctx context.Context, appointmentID string, payerID uuid.UUID) (*apppayment.PaymentEligibility, error) {
	endpoint := fmt.Sprintf("%s/internal/appointments/%s/payment-eligibility", c.baseURL, url.PathEscape(appointmentID))
	body, err := json.Marshal(paymentEligibilityRequest{PayerID: payerID.String()})
	if err != nil {
		return nil, fmt.Errorf("booking_client: marshal eligibility request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("booking_client: create eligibility request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("booking_client: get payment eligibility %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapPaymentEligibilityStatus(resp)
	}

	var apiResp paymentEligibilityAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("booking_client: decode eligibility response: %w", err)
	}

	return &apppayment.PaymentEligibility{
		AppointmentID: apiResp.Data.AppointmentID,
		ExpertID:      apiResp.Data.ExpertID,
		AmountVND:     apiResp.Data.AmountVND,
		ExpiresAt:     apiResp.Data.ExpiresAt,
	}, nil
}

func mapPaymentEligibilityStatus(resp *http.Response) error {
	respBody, _ := io.ReadAll(resp.Body)
	detail := fmt.Sprintf("booking_client: payment eligibility status %d: %s", resp.StatusCode, string(respBody))

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", apppayment.ErrAppointmentNotFound, detail)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", apppayment.ErrPaymentEligibilityForbidden, detail)
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", apppayment.ErrPaymentEligibilityConflict, detail)
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", apppayment.ErrInvalidBookingPrice, detail)
	default:
		return fmt.Errorf("%s", detail)
	}
}

func (c *restBookingClient) ConfirmAppointment(ctx context.Context, appointmentID string) error {
	return c.callWebhook(ctx, appointmentID, "SUCCESS")
}

func (c *restBookingClient) FailAppointment(ctx context.Context, appointmentID string) error {
	return c.callWebhook(ctx, appointmentID, "FAILED")
}

func (c *restBookingClient) callWebhook(ctx context.Context, appointmentID, status string) error {
	endpoint := fmt.Sprintf("%s/internal/appointments/%s/webhook", c.baseURL, url.PathEscape(appointmentID))

	payload := webhookPayload{
		AppointmentID: appointmentID,
		Status:        status,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("booking_client: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("booking_client: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("booking_client: call %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("booking_client: unexpected status %d from booking-service: %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[booking_client] appointment %s -> %s accepted by booking-service", appointmentID, status)
	return nil
}
