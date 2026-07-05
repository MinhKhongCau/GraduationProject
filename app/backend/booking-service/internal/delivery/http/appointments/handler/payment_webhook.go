package handler

import (
	"booking-service/internal/domain"
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PaymentWebhookRequest - Dữ liệu cổng thanh toán gửi về
type PaymentWebhookRequest struct {
	AppointmentID string `json:"appointment_id" binding:"required"`
	Status        string `json:"status" binding:"required"` // "SUCCESS" hoặc "FAILED"
}

type PaymentWebhookHandler struct {
	repo *postgres.AppointmentRepository
}

func NewPaymentWebhookHandler(repo *postgres.AppointmentRepository) *PaymentWebhookHandler {
	return &PaymentWebhookHandler{repo: repo}
}

// Handle - POST /api/v1/appointments/webhook
//
//	@Summary      Nhận kết quả thanh toán từ cổng thanh toán (Webhook)
//	@Description  Cổng thanh toán gọi endpoint này để thông báo kết quả. SUCCESS -> CONFIRMED, FAILED -> CANCELLED
//	@Tags         Appointments
//	@Accept       json
//	@Produce      json
//	@Param        body  body      PaymentWebhookRequest  true  "Kết quả thanh toán"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /appointments/webhook [post]
func (h *PaymentWebhookHandler) Handle(c *gin.Context) {
	var req PaymentWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu webhook không hợp lệ", err.Error())
		return
	}

	if req.Status == "SUCCESS" {
		// Thanh toán thành công: Chuyển appointment -> CONFIRMED, slot -> OCCUPIED
		if err := h.repo.ConfirmPayment(req.AppointmentID); err != nil {
			response.Error(c, http.StatusInternalServerError, "Lỗi xác nhận thanh toán", err.Error())
			return
		}
		response.Success(c, "Xác nhận thanh toán thành công! Lịch hẹn đã được xác nhận.", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusConfirmed,
		})
	} else {
		// Thanh toán thất bại: Mở lock tức thì (Worker sẽ dọn định kỳ nếu miss)
		response.Success(c, "Đã nhận thông báo thanh toán thất bại. Slot sẽ được mở lại sau ít phút.", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusCancelled,
		})
	}
}
