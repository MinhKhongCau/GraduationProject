package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apppayment "payment-service/internal/payment/application"
	paymentdomain "payment-service/internal/payment/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestListCompensationCasesRequiresAdminAndTrustedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name   string
		role   string
		userID string
		status int
	}{
		{name: "non admin", role: "PATIENT", userID: uuid.NewString(), status: http.StatusForbidden},
		{name: "missing identity", role: "ADMIN", status: http.StatusUnauthorized},
		{name: "invalid identity", role: "ADMIN", userID: "invalid", status: http.StatusUnauthorized},
	} {
		t.Run(test.name, func(t *testing.T) {
			usecase := &fakeCreateOrderUsecase{}
			response := serveCompensationRequest(usecase, http.MethodGet, "/api/v1/payments/compensation-cases", test.role, test.userID)
			if response.Code != test.status || usecase.listCalled {
				t.Fatalf("unexpected authorization result: status=%d called=%v", response.Code, usecase.listCalled)
			}
		})
	}
}

func TestListCompensationCasesPassesFiltersAndReturnsSafeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	appointmentID, orderID, caseID := uuid.New(), uuid.New(), uuid.New()
	usecase := &fakeCreateOrderUsecase{compensationPage: &apppayment.CompensationCasePage{
		Items: []apppayment.CompensationCase{{
			ID: caseID, PaymentOrderID: orderID, AppointmentID: appointmentID,
			Type: paymentdomain.CompensationBookingFulfillment, Status: paymentdomain.CompensationRefundRequired,
			ReasonCode: paymentdomain.CompensationReasonBookingConflict, SafeReason: "safe", AmountVND: 250000,
			GatewayOrderReference: orderID.String(), GatewayTransactionNumber: "vnp-transaction-number",
			GatewayResponseCode: "00", GatewayTransactionStatus: "00", GatewayPaymentDate: "20260719100000",
		}},
		Total: 1, Page: 2, Size: 10,
	}}
	path := "/api/v1/payments/compensation-cases?status=REFUND_REQUIRED&appointment_id=" + appointmentID.String() + "&payment_order_id=" + orderID.String() + "&page=2&size=10"
	response := serveCompensationRequest(usecase, http.MethodGet, path, "ADMIN", uuid.NewString())
	if response.Code != http.StatusOK || !usecase.listCalled {
		t.Fatalf("unexpected list response: status=%d body=%s", response.Code, response.Body.String())
	}
	filter := usecase.compensationFilter
	if filter.Status != paymentdomain.CompensationRefundRequired || filter.AppointmentID == nil || *filter.AppointmentID != appointmentID || filter.PaymentOrderID == nil || *filter.PaymentOrderID != orderID || filter.Page != 2 || filter.Size != 10 {
		t.Fatalf("filters were not preserved: %+v", filter)
	}
	if !strings.Contains(response.Body.String(), "vnp-transaction-number") || strings.Contains(response.Body.String(), "secure_hash") || strings.Contains(response.Body.String(), "raw_query") || strings.Contains(response.Body.String(), "authorization") {
		t.Fatalf("sensitive gateway fields leaked: %s", response.Body.String())
	}
}

func TestGetCompensationCaseReturnsDetailAndNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	caseID := uuid.New()
	usecase := &fakeCreateOrderUsecase{compensationCase: &apppayment.CompensationCase{ID: caseID}}
	response := serveCompensationRequest(usecase, http.MethodGet, "/api/v1/payments/compensation-cases/"+caseID.String(), "ADMIN", uuid.NewString())
	if response.Code != http.StatusOK || !usecase.getCalled || usecase.compensationCaseID != caseID {
		t.Fatalf("unexpected detail response: status=%d body=%s", response.Code, response.Body.String())
	}

	usecase = &fakeCreateOrderUsecase{compensationErr: apppayment.ErrCompensationCaseNotFound}
	response = serveCompensationRequest(usecase, http.MethodGet, "/api/v1/payments/compensation-cases/"+caseID.String(), "ADMIN", uuid.NewString())
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", response.Code, response.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body["success"] != false {
		t.Fatalf("unexpected error contract: %s", response.Body.String())
	}
}

func serveCompensationRequest(usecase apppayment.Usecase, method, path, role, userID string) *httptest.ResponseRecorder {
	router := gin.New()
	handler := NewHandler(usecase)
	router.GET("/api/v1/payments/compensation-cases", handler.ListCompensationCases)
	router.GET("/api/v1/payments/compensation-cases/:id", handler.GetCompensationCase)
	request := httptest.NewRequest(method, path, nil)
	if role != "" {
		request.Header.Set("X-User-Role", role)
	}
	if userID != "" {
		request.Header.Set("X-User-Id", userID)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
