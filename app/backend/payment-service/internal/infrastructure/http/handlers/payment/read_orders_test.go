package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/application/readquery"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (u *fakeCreateOrderUsecase) ListPaymentOrders(context.Context, apppayment.PaymentOrderFilter) (*readquery.Page[apppayment.PaymentOrderView], error) {
	page := readquery.NewPage([]apppayment.PaymentOrderView{}, readquery.PageRequest{Page: 0, Size: 20}, 0)
	return &page, nil
}
func (u *fakeCreateOrderUsecase) GetPaymentOrder(context.Context, uuid.UUID, uuid.UUID, bool) (*apppayment.PaymentOrderView, error) {
	return nil, apppayment.ErrPaymentOrderForbidden
}

func TestPaymentOrderListRejectsInvalidPaginationAndMissingIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/payments/orders", NewHandler(&fakeCreateOrderUsecase{}).ListPaymentOrders)
	tests := []struct {
		name, query, role, user string
		status                  int
	}{{name: "negative page", query: "?page=-1", role: "PATIENT", user: uuid.NewString(), status: http.StatusBadRequest}, {name: "oversized", query: "?size=101", role: "PATIENT", user: uuid.NewString(), status: http.StatusBadRequest}, {name: "missing identity", role: "PATIENT", status: http.StatusUnauthorized}, {name: "wrong role", role: "EXPERT", user: uuid.NewString(), status: http.StatusForbidden}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/payments/orders"+tt.query, nil)
			request.Header.Set("X-User-Role", tt.role)
			request.Header.Set("X-User-Id", tt.user)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.status {
				t.Fatalf("expected %d, got %d: %s", tt.status, recorder.Code, recorder.Body.String())
			}
		})
	}
}
