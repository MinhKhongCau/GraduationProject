package handler

import (
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
