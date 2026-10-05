package payment

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVNPayIPNAlreadyProcessedReturnsStableAcknowledgement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := &fakeCreateOrderUsecase{processAlready: true}
	router := gin.New()
	router.GET("/api/v1/payments/vnpay-ipn", NewHandler(usecase).HandleVNPayIPN)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/payments/vnpay-ipn", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"RspCode":"02"`) {
		t.Fatalf("unexpected idempotent IPN response: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestVNPayIPNMapsSignatureAndAmountErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		rspCode string
		message string
	}{
		{name: "missing secure hash", err: errors.New("missing vnp_SecureHash"), rspCode: "97", message: "Invalid Signature"},
		{name: "invalid mixed-case amount", err: errors.New("invalid vnp_Amount: invalid syntax"), rspCode: "04", message: "Invalid amount"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			usecase := &fakeCreateOrderUsecase{processErr: tt.err}
			router := gin.New()
			router.GET("/api/v1/payments/vnpay-ipn", NewHandler(usecase).HandleVNPayIPN)

			request := httptest.NewRequest(http.MethodGet, "/api/v1/payments/vnpay-ipn", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"RspCode":"`+tt.rspCode+`"`) || !strings.Contains(response.Body.String(), `"Message":"`+tt.message+`"`) {
				t.Fatalf("unexpected IPN error response: status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
