package payment

import (
	"errors"
	"net/http"
	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/infrastructure/http/middleware"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SettlementCallerID là service duy nhất được phép yêu cầu chi trả thù lao.
const SettlementCallerID = "booking-service"

// SettleSession handles POST /internal/payments/appointments/:id/settle
// Booking-service gọi sau khi buổi tư vấn hoàn tất để chuyển net_amount vào ví chuyên gia.
func (h *Handler) SettleSession(c *gin.Context) {
	if h.settlement == nil {
		response.Error(c, http.StatusNotImplemented, "Session settlement is not configured", "not_implemented")
		return
	}
	if caller, _ := middleware.GetCallerID(c); caller != SettlementCallerID {
		response.Error(c, http.StatusForbidden, "Caller is not allowed to settle sessions", "forbidden")
		return
	}
	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid appointment ID", "invalid_appointment_id")
		return
	}

	settlement, err := h.settlement.SettleCompletedSession(c.Request.Context(), appointmentID)
	switch {
	case err == nil:
		response.Success(c, "Session settled", settlement)
	case errors.Is(err, apppayment.ErrSessionOrderNotFound):
		response.Error(c, http.StatusNotFound, "No successful payment for appointment", "payment_not_found")
	case errors.Is(err, apppayment.ErrSessionNotSettleable):
		response.Error(c, http.StatusConflict, "Payment is under review or refund", "payment_not_settleable")
	default:
		response.Error(c, http.StatusInternalServerError, "Failed to settle session", err.Error())
	}
}
