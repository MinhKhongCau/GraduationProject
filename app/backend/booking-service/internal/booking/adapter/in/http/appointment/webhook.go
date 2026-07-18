package handler

import (
	"booking-service/internal/booking/domain"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebhookRequest - Dữ liệu cổng thanh toán gửi về
type WebhookRequest struct {
	AppointmentID string `json:"appointment_id" binding:"required"`
	Status        string `json:"status" binding:"required"` // "SUCCESS" hoặc "FAILED"
}

// Webhook - POST /api/v1/appointments/webhook
//
//	@Summary      [SYSTEM/WEBHOOK] Receive payment webhook
//	@Description  [SYSTEM/WEBHOOK] Payment gateway calls this endpoint to notify status. SUCCESS -> CONFIRMED, FAILED -> CANCELLED
//	@Tags         Appointments
//	@Accept       json
//	@Produce      json
//	@Param        body  body      WebhookRequest  true  "Kết quả thanh toán"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /public/booking/appointments/webhook [post]
func (h *Handler) Webhook(c *gin.Context) {
	// TODO: VERIFY HMAC SIGNATURE từ VNPay/Momo tại đây
	// Bắt buộc verify chữ ký để tránh giả mạo webhook confirm thanh toán.

	// TODO ỨNG DỤNG REDIS NHA: Sử dụng Idempotency Key
	// Lưu transaction id (hoặc order_id) vào Redis để tránh xử lý trùng lặp nếu Webhook gọi lại nhiều lần.

	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid webhook data", err.Error())
		return
	}

	if req.Status == "SUCCESS" {
		// Successful payment: Transition appointment -> CONFIRMED, slot -> OCCUPIED
		if err := h.usecase.ConfirmPayment(req.AppointmentID); err != nil {
			response.Error(c, http.StatusInternalServerError, "Payment confirmation failed", err.Error())
			return
		}
		response.Success(c, "Payment confirmed successfully! Appointment is confirmed.", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusConfirmed,
		})
	} else {
		// Failed payment
		_ = h.usecase.HandlePaymentFailure(req.AppointmentID)
		response.Success(c, "Received payment failure notification. The slot will be unlocked shortly.", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusCancelled,
		})
	}
}
