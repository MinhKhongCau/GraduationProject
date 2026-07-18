package handler

import (
	"booking-service/internal/appointment"
	"booking-service/internal/domain"
	"booking-service/pkg/internal_auth"
	"booking-service/pkg/response"
	"errors"
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
	if !isPaymentServiceCaller(c) {
		response.Error(c, http.StatusForbidden, "Forbidden", "caller is not allowed to update payment result")
		return
	}

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

	status, err := appointment.ParsePaymentResultStatus(req.Status)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment result status", err.Error())
		return
	}

	if err := h.usecase.HandlePaymentResult(appointment.HandlePaymentResultCommand{
		AppointmentID: req.AppointmentID,
		Status:        status,
	}); err != nil {
		respondPaymentResultError(c, err)
		return
	}

	if status == appointment.PaymentResultSuccess {
		response.Success(c, "Appointment confirmed successfully", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusConfirmed,
		})
	} else {
		response.Success(c, "Appointment cancelled due to payment failure", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusCancelled,
		})
	}
}

// InternalGetAppointment - GET /internal/appointments/:id
// Được gọi bởi Payment Service để lấy thông tin xác minh đơn hàng.
// Bảo vệ bởi internal_auth.Middleware() — chỉ service nội bộ mới gọi được.
//
//	@Summary      [INTERNAL] Lấy chi tiết lịch hẹn
//	@Description  [INTERNAL] Payment Service gọi endpoint này để lấy chi tiết lịch hẹn (verify trước khi tạo Order).
//	@Tags         Internal
//	@Produce      json
//	@Param        id    path      string  true  "Appointment ID"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      404   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /internal/appointments/{id} [get]
func (h *Handler) InternalGetAppointment(c *gin.Context) {
	if !isPaymentServiceCaller(c) {
		response.Error(c, http.StatusForbidden, "Forbidden", "caller is not allowed to read payment appointment data")
		return
	}

	appointmentID := c.Param("id")
	if appointmentID == "" {
		response.Error(c, http.StatusBadRequest, "appointment_id is required", "missing path param :id")
		return
	}

	appt, err := h.usecase.GetAppointmentByID(appointmentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Appointment not found", err.Error())
		return
	}

	response.Success(c, "Get appointment successfully", appt)
}

func isPaymentServiceCaller(c *gin.Context) bool {
	callerID, ok := internal_auth.GetCallerID(c)
	return ok && callerID == "payment-service"
}

func respondPaymentResultError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, appointment.ErrInvalidPaymentResultStatus):
		response.Error(c, http.StatusBadRequest, "Invalid payment result status", err.Error())
	case errors.Is(err, appointment.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Appointment not found", err.Error())
	case errors.Is(err, appointment.ErrInvalidStatus), errors.Is(err, appointment.ErrPaymentResultConflict):
		response.Error(c, http.StatusConflict, "Payment result conflicts with booking state", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Payment result update failed", err.Error())
	}
}
