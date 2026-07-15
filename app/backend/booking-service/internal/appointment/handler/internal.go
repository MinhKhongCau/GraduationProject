package handler

import (
	"booking-service/internal/domain"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ConfirmInternal — POST /internal/appointments/:id/webhook
// Được gọi bởi Payment Service sau khi thanh toán thành công.
// Bảo vệ bởi internal_auth.Middleware() — chỉ service nội bộ mới gọi được.
//
//	@Summary      [INTERNAL] Nhận kết quả thanh toán từ Payment Service
//	@Description  [INTERNAL] Payment Service gọi endpoint này sau khi xử lý IPN từ VNPay. Chuyển appointment → CONFIRMED hoặc CANCELLED.
//	@Tags         Internal
//	@Accept       json
//	@Produce      json
//	@Param        id    path      string          true  "Appointment ID"
//	@Param        body  body      WebhookRequest  true  "Kết quả thanh toán"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /internal/appointments/{id}/webhook [post]
func (h *Handler) InternalPaymentWebhook(c *gin.Context) {
	appointmentID := c.Param("id")
	if appointmentID == "" {
		response.Error(c, http.StatusBadRequest, "appointment_id is required", "missing path param :id")
		return
	}

	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid webhook payload", err.Error())
		return
	}

	// Override appointment_id từ path param để đảm bảo nhất quán
	// (Payment Service gửi cả 2: path param và body — lấy path param làm chuẩn)
	req.AppointmentID = appointmentID

	if req.Status == "SUCCESS" {
		// Idempotency check: nếu appointment đã CONFIRMED rồi thì không làm gì cả.
		// Logic này nằm trong usecase.ConfirmPayment → repository.
		if err := h.usecase.ConfirmPayment(req.AppointmentID); err != nil {
			response.Error(c, http.StatusInternalServerError, "Payment confirmation failed", err.Error())
			return
		}
		response.Success(c, "Appointment confirmed successfully", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusConfirmed,
		})
	} else {
		// FAILED: hủy lịch hẹn, giải phóng slot
		_ = h.usecase.HandlePaymentFailure(req.AppointmentID)
		response.Success(c, "Appointment cancelled due to payment failure", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusCancelled,
		})
	}
}
