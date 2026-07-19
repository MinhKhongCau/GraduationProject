package bookingrest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	apppayment "payment-service/internal/payment/application"
	"payment-service/pkg/httpclient"

	"github.com/google/uuid"
)

type restBookingClient struct {
	baseURL         string
	eligibilityHTTP *httpclient.Client
	deliveryHTTP    *httpclient.Client
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

type webhookAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewRestBookingClient(baseURL string, tokenProvider httpclient.TokenProvider) apppayment.BookingServiceClient {
	return newRestBookingClient(
		baseURL,
		httpclient.New(httpclient.Options{
			MaxRetries:    3,
			TokenProvider: tokenProvider,
		}),
		httpclient.New(httpclient.Options{
			DisableRetries: true,
			TokenProvider:  tokenProvider,
		}),
	)
}

func newRestBookingClient(baseURL string, eligibilityHTTP, deliveryHTTP *httpclient.Client) *restBookingClient {
	return &restBookingClient{baseURL: baseURL, eligibilityHTTP: eligibilityHTTP, deliveryHTTP: deliveryHTTP}
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

	resp, err := c.eligibilityHTTP.Do(ctx, req)
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
	body, err := json.Marshal(webhookPayload{AppointmentID: appointmentID, Status: status})
	if err != nil {
		return permanentDeliveryError(apppayment.BookingDeliveryBadRequest, "booking webhook payload could not be encoded")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return permanentDeliveryError(apppayment.BookingDeliveryBadRequest, "booking webhook request could not be created")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.deliveryHTTP.Do(ctx, req)
	if err != nil {
		category := apppayment.BookingDeliveryNetwork
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			category = apppayment.BookingDeliveryTimeout
		}
		return retryableDeliveryError(category, "booking webhook request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}
		var apiResp webhookAPIResponse
		if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&apiResp); err != nil {
			return retryableDeliveryError(apppayment.BookingDeliveryMalformedResponse, "booking webhook returned a malformed success response")
		}
		if !apiResp.Success {
			return retryableDeliveryError(apppayment.BookingDeliveryMalformedResponse, "booking webhook returned an ambiguous unsuccessful 2xx response")
		}
		log.Printf("[booking_client] appointment %s -> %s accepted by booking-service", appointmentID, status)
		return nil
	}

	return classifyWebhookStatus(resp.StatusCode)
}

func classifyWebhookStatus(statusCode int) error {
	switch statusCode {
	case http.StatusRequestTimeout:
		return retryableDeliveryError(apppayment.BookingDeliveryTimeout, "booking webhook returned HTTP 408")
	case http.StatusTooManyRequests:
		return retryableDeliveryError(apppayment.BookingDeliveryRateLimited, "booking webhook returned HTTP 429")
	case http.StatusUnauthorized, http.StatusForbidden:
		return permanentDeliveryError(apppayment.BookingDeliveryAuthentication, fmt.Sprintf("booking webhook returned HTTP %d", statusCode))
	case http.StatusNotFound:
		return permanentDeliveryError(apppayment.BookingDeliveryNotFound, "booking webhook returned HTTP 404")
	case http.StatusConflict:
		return permanentDeliveryError(apppayment.BookingDeliveryConflict, "booking webhook returned HTTP 409")
	case http.StatusBadRequest:
		return permanentDeliveryError(apppayment.BookingDeliveryBadRequest, "booking webhook returned HTTP 400")
	default:
		if statusCode >= 500 && statusCode <= 599 {
			return retryableDeliveryError(apppayment.BookingDeliveryUpstream, fmt.Sprintf("booking webhook returned HTTP %d", statusCode))
		}
		return permanentDeliveryError(apppayment.BookingDeliveryBusinessRejection, fmt.Sprintf("booking webhook returned unexpected HTTP %d", statusCode))
	}
}

func retryableDeliveryError(category apppayment.BookingDeliveryFailureCategory, message string) error {
	return &apppayment.BookingDeliveryError{Category: category, Retryable: true, Message: message}
}

func permanentDeliveryError(category apppayment.BookingDeliveryFailureCategory, message string) error {
	return &apppayment.BookingDeliveryError{Category: category, Retryable: false, Message: message}
}
