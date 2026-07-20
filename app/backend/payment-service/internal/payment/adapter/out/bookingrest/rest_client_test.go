package bookingrest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	apppayment "payment-service/internal/payment/application"
	"payment-service/pkg/httpclient"
	"sync/atomic"
	"testing"
	"time"
)

type staticTokenProvider struct {
	token       string
	invalidated atomic.Int32
}

func (p *staticTokenProvider) GetToken(context.Context) (string, error) { return p.token, nil }
func (p *staticTokenProvider) InvalidateToken()                         { p.invalidated.Add(1) }

func TestBookingWebhookSuccessUsesCurrentContractAndOneRequest(t *testing.T) {
	appointmentID := "b76308df-049d-41e0-a18c-ab2fce34482d"
	tokenProvider := &staticTokenProvider{token: "internal-token"}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/internal/appointments/"+appointmentID+"/webhook" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer internal-token" {
			t.Fatalf("unexpected authorization header %q", r.Header.Get("Authorization"))
		}
		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.AppointmentID != appointmentID || payload.Status != "SUCCESS" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"Appointment confirmed successfully","data":{"status":1}}`))
	}))
	defer server.Close()

	client := NewRestBookingClient(server.URL, tokenProvider)
	if err := client.ConfirmAppointment(context.Background(), appointmentID); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one request, got %d", calls.Load())
	}
}

func TestBookingWebhookIdempotentNoopResponseIsDelivered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"message":"Appointment confirmed successfully"}`))
	}))
	defer server.Close()
	if err := NewRestBookingClient(server.URL, nil).ConfirmAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d"); err != nil {
		t.Fatal(err)
	}
}

func TestBookingWebhookIdempotentFailedNoopResponseIsDelivered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"message":"Appointment cancelled due to payment failure","data":{"status":2}}`))
	}))
	defer server.Close()
	if err := NewRestBookingClient(server.URL, nil).FailAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d"); err != nil {
		t.Fatal(err)
	}
}

func TestBookingWebhookFailurePayloadRemainsUnchanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload webhookPayload
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.Status != "FAILED" {
			t.Fatalf("expected FAILED, got %q", payload.Status)
		}
		_, _ = w.Write([]byte(`{"success":true,"message":"Appointment cancelled due to payment failure"}`))
	}))
	defer server.Close()
	if err := NewRestBookingClient(server.URL, nil).FailAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d"); err != nil {
		t.Fatal(err)
	}
}

func TestBookingWebhookStatusClassificationAndNoInternalRetry(t *testing.T) {
	tests := []struct {
		status    int
		category  apppayment.BookingDeliveryFailureCategory
		retryable bool
	}{
		{http.StatusBadRequest, apppayment.BookingDeliveryBadRequest, false},
		{http.StatusUnauthorized, apppayment.BookingDeliveryAuthentication, false},
		{http.StatusForbidden, apppayment.BookingDeliveryAuthentication, false},
		{http.StatusNotFound, apppayment.BookingDeliveryNotFound, false},
		{http.StatusConflict, apppayment.BookingDeliveryConflict, false},
		{http.StatusRequestTimeout, apppayment.BookingDeliveryTimeout, true},
		{http.StatusTooManyRequests, apppayment.BookingDeliveryRateLimited, true},
		{http.StatusInternalServerError, apppayment.BookingDeliveryUpstream, true},
		{http.StatusBadGateway, apppayment.BookingDeliveryUpstream, true},
		{http.StatusServiceUnavailable, apppayment.BookingDeliveryUpstream, true},
		{http.StatusGatewayTimeout, apppayment.BookingDeliveryUpstream, true},
	}
	for _, test := range tests {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(`{"success":false,"error":"internal detail is not persisted"}`))
			}))
			defer server.Close()

			err := NewRestBookingClient(server.URL, nil).ConfirmAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d")
			var deliveryErr *apppayment.BookingDeliveryError
			if !errors.As(err, &deliveryErr) {
				t.Fatalf("expected typed delivery error, got %v", err)
			}
			if deliveryErr.Category != test.category || deliveryErr.Retryable != test.retryable {
				t.Fatalf("unexpected classification: %+v", deliveryErr)
			}
			if calls.Load() != 1 {
				t.Fatalf("expected one request, got %d", calls.Load())
			}
		})
	}
}

func TestBookingWebhookMalformedSuccessResponseIsRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()
	err := NewRestBookingClient(server.URL, nil).ConfirmAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d")
	assertDeliveryError(t, err, apppayment.BookingDeliveryMalformedResponse, true)
}

func TestBookingWebhookTwoHundredBusinessFailureIsAmbiguousAndRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"message":"rejected"}`))
	}))
	defer server.Close()
	err := NewRestBookingClient(server.URL, nil).ConfirmAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d")
	assertDeliveryError(t, err, apppayment.BookingDeliveryMalformedResponse, true)
}

func TestBookingPaymentResultContractClassification(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		category   apppayment.BookingDeliveryFailureCategory
		retryable  bool
		delivered  bool
	}{
		{name: "successful transition", statusCode: http.StatusOK, body: `{"success":true,"message":"Appointment confirmed successfully","data":{"status":1}}`, delivered: true},
		{name: "idempotent no-op", statusCode: http.StatusOK, body: `{"success":true,"message":"Appointment confirmed successfully","data":{"status":1}}`, delivered: true},
		{name: "appointment not found", statusCode: http.StatusNotFound, body: `{"success":false,"message":"Appointment not found","error":"appointment not found"}`, category: apppayment.BookingDeliveryNotFound},
		{name: "invalid status payload", statusCode: http.StatusBadRequest, body: `{"success":false,"message":"Invalid payment result status","error":"invalid payment result status"}`, category: apppayment.BookingDeliveryBadRequest},
		{name: "caller authentication failure", statusCode: http.StatusForbidden, body: `{"success":false,"message":"Forbidden","error":"caller is not allowed to update payment result"}`, category: apppayment.BookingDeliveryAuthentication},
		{name: "cancelled plus success conflict", statusCode: http.StatusConflict, body: `{"success":false,"message":"Payment result conflicts with booking state","error":"cancelled appointment cannot be confirmed"}`, category: apppayment.BookingDeliveryConflict},
		{name: "confirmed plus failed conflict", statusCode: http.StatusConflict, body: `{"success":false,"message":"Payment result conflicts with booking state","error":"confirmed appointment cannot be cancelled by payment failure"}`, category: apppayment.BookingDeliveryConflict},
		{name: "repository failure", statusCode: http.StatusInternalServerError, body: `{"success":false,"message":"Payment result update failed","error":"database failure"}`, category: apppayment.BookingDeliveryUpstream, retryable: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			err := NewRestBookingClient(server.URL, nil).ConfirmAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d")
			if test.delivered {
				if err != nil {
					t.Fatalf("expected delivered response, got %v", err)
				}
				return
			}
			assertDeliveryError(t, err, test.category, test.retryable)
		})
	}
}

func TestBookingWebhookAuthenticationFailureInvalidatesTokenWithoutRetry(t *testing.T) {
	provider := &staticTokenProvider{token: "static-token"}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	err := NewRestBookingClient(server.URL, provider).ConfirmAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d")
	assertDeliveryError(t, err, apppayment.BookingDeliveryAuthentication, false)
	if calls.Load() != 1 || provider.invalidated.Load() != 1 {
		t.Fatalf("expected one request and token invalidation, calls=%d invalidations=%d", calls.Load(), provider.invalidated.Load())
	}
}

func TestBookingWebhookTimeoutIsRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	client := newRestBookingClient(server.URL, nil, httpclient.New(httpclient.Options{Timeout: 5 * time.Millisecond, DisableRetries: true}))
	err := client.ConfirmAppointment(context.Background(), "b76308df-049d-41e0-a18c-ab2fce34482d")
	assertDeliveryError(t, err, apppayment.BookingDeliveryTimeout, true)
}

func assertDeliveryError(t *testing.T, err error, category apppayment.BookingDeliveryFailureCategory, retryable bool) {
	t.Helper()
	var deliveryErr *apppayment.BookingDeliveryError
	if !errors.As(err, &deliveryErr) {
		t.Fatalf("expected typed delivery error, got %v", err)
	}
	if deliveryErr.Category != category || deliveryErr.Retryable != retryable {
		t.Fatalf("unexpected classification: %+v", deliveryErr)
	}
}
