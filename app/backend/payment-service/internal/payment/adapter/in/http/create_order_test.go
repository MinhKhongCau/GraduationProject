package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	apppayment "payment-service/internal/payment/application"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestCreateOrderUsesTrustedHeaderAndIgnoresBodyPayerID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	trustedPayerID := uuid.New()
	bodyPayerID := uuid.New()
	expertID := uuid.New()
	appointmentID := uuid.New()
	usecase := &fakeCreateOrderUsecase{
		order: &entity.PaymentOrder{
			ID:               uuid.New(),
			PayerID:          trustedPayerID,
			ExpertID:         expertID,
			GrossAmount:      vo.Money(100000),
			NetAmount:        vo.Money(85000),
			CommissionAmount: vo.Money(15000),
			Gateway:          "VNPAY",
			Status:           entity.OrderStatusPending,
			AppointmentID:    &appointmentID,
		},
		paymentURL: "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html",
	}

	body := []byte(`{
		"payer_id":"` + bodyPayerID.String() + `",
		"expert_id":"` + expertID.String() + `",
		"amount":100000,
		"gateway":"MOMO",
		"appointment_id":"` + appointmentID.String() + `"
	}`)
	res := performCreateOrderRequest(usecase, trustedPayerID.String(), body)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if !usecase.createCalled {
		t.Fatal("expected usecase to be called")
	}
	if usecase.payerID != trustedPayerID {
		t.Fatalf("expected trusted header payer %s, got %s", trustedPayerID, usecase.payerID)
	}
	if usecase.appointmentID != appointmentID.String() {
		t.Fatalf("expected appointment_id %s, got %s", appointmentID, usecase.appointmentID)
	}
}

func TestCreateOrderRejectsMissingTrustedPayerEvenWhenBodyContainsPayerID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expertID := uuid.New()
	bodyPayerID := uuid.New()
	appointmentID := uuid.New()
	usecase := &fakeCreateOrderUsecase{}
	body := []byte(`{
		"payer_id":"` + bodyPayerID.String() + `",
		"expert_id":"` + expertID.String() + `",
		"amount":100000,
		"gateway":"VNPAY",
		"appointment_id":"` + appointmentID.String() + `"
	}`)
	res := performCreateOrderRequest(usecase, "", body)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", res.Code, res.Body.String())
	}
	if usecase.createCalled {
		t.Fatal("expected usecase not to be called")
	}
}

func TestCreateOrderRejectsInvalidTrustedPayer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	usecase := &fakeCreateOrderUsecase{}
	body := []byte(`{"appointment_id":"` + uuid.New().String() + `"}`)
	res := performCreateOrderRequest(usecase, "not-a-uuid", body)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", res.Code, res.Body.String())
	}
	if usecase.createCalled {
		t.Fatal("expected usecase not to be called")
	}
}

func TestCreateOrderRejectsMissingOrEmptyAppointmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body []byte
	}{
		{name: "missing", body: []byte(`{}`)},
		{name: "legacy fields only", body: []byte(`{"payer_id":"` + uuid.New().String() + `","expert_id":"` + uuid.New().String() + `","amount":100000,"gateway":"VNPAY"}`)},
		{name: "null", body: []byte(`{"appointment_id":null}`)},
		{name: "empty", body: []byte(`{"appointment_id":""}`)},
		{name: "whitespace", body: []byte(`{"appointment_id":"   "}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeCreateOrderUsecase{}
			res := performCreateOrderRequest(usecase, uuid.New().String(), tt.body)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
			}
			if usecase.createCalled {
				t.Fatal("expected usecase not to be called")
			}
		})
	}
}

func TestCreateOrderMapsUsecaseErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "unsupported gateway", err: apppayment.ErrUnsupportedGateway, wantStatus: http.StatusBadRequest},
		{name: "invalid request", err: apppayment.ErrInvalidCreateOrderRequest, wantStatus: http.StatusBadRequest},
		{name: "invalid booking data", err: apppayment.ErrInvalidBookingData, wantStatus: http.StatusBadRequest},
		{name: "ownership", err: apppayment.ErrAppointmentOwnership, wantStatus: http.StatusForbidden},
		{name: "booking not found", err: apppayment.ErrBookingAppointmentNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid state", err: apppayment.ErrAppointmentInvalidState, wantStatus: http.StatusConflict},
		{name: "unexpected", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeCreateOrderUsecase{err: tt.err}
			body := []byte(`{"appointment_id":"` + uuid.New().String() + `"}`)
			res := performCreateOrderRequest(usecase, uuid.New().String(), body)

			if res.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, res.Code, res.Body.String())
			}
			if !usecase.createCalled {
				t.Fatal("expected usecase to be called")
			}
		})
	}
}

func performCreateOrderRequest(usecase *fakeCreateOrderUsecase, payerID string, body []byte) *httptest.ResponseRecorder {
	router := gin.New()
	router.POST("/api/v1/payments/orders", NewHandler(usecase).CreateOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if payerID != "" {
		req.Header.Set("X-User-Id", payerID)
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

type fakeCreateOrderUsecase struct {
	createCalled  bool
	payerID       uuid.UUID
	appointmentID string
	order         *entity.PaymentOrder
	paymentURL    string
	err           error
}

func (u *fakeCreateOrderUsecase) CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, ipAddr string) (*entity.PaymentOrder, string, error) {
	u.createCalled = true
	u.payerID = payerID
	u.appointmentID = appointmentID
	if u.err != nil {
		return nil, "", u.err
	}
	return u.order, u.paymentURL, nil
}

func (u *fakeCreateOrderUsecase) ProcessIPN(ctx context.Context, params map[string][]string) (bool, error) {
	return false, nil
}
