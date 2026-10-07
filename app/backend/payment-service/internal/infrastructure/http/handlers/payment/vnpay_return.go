package payment

import (
	"errors"
	"net/http"
	"strings"

	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
)

// HandleVNPayReturn handles GET /api/v1/payments/vnpay-return
// @Summary      [PUBLIC] Verify the VNPay browser return and get payment + booking status
// @Description  The FE payment-result page forwards the full VNPay query string it received on vnp_ReturnUrl. The service verifies vnp_SecureHash, settles the order idempotently (same logic as the IPN) and returns the payment status plus the current booking status. Role: Public (authenticated by the VNPay checksum).
// @Tags         Payment Orders
// @Produce      json
// @Param        queryParams  query     object  false  "VNPay return parameters (vnp_*)"
// @Success      200          {object}  response.Response{data=apppayment.PaymentReturnResult}
// @Failure      400          {object}  response.Response
// @Failure      404          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/vnpay-return [get]
func (h *Handler) HandleVNPayReturn(c *gin.Context) {
	returnUsecase, ok := h.usecase.(apppayment.ReturnUsecase)
	if !ok {
		response.Error(c, http.StatusServiceUnavailable, "Payment return is unavailable", "return usecase is not configured")
		return
	}

	result, err := returnUsecase.ProcessReturn(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		classification := strings.ToLower(err.Error())
		switch {
		case strings.Contains(classification, "checksum") || strings.Contains(classification, "securehash"):
			response.Error(c, http.StatusBadRequest, "Invalid VNPay signature", err.Error())
		case errors.Is(err, apppayment.ErrPaymentOrderNotFound) || strings.Contains(classification, "not found"):
			response.Error(c, http.StatusNotFound, "Payment order not found", err.Error())
		case strings.Contains(classification, "amount") || strings.Contains(classification, "missing vnp_") || strings.Contains(classification, "invalid order"):
			response.Error(c, http.StatusBadRequest, "Invalid VNPay return parameters", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Could not process VNPay return", err.Error())
		}
		return
	}

	response.Success(c, "Payment status retrieved", result)
}
