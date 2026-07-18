package handler

import (
	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/domain"
	"booking-service/pkg/internal_auth"
	"booking-service/pkg/response"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type paymentEligibilityRequest struct {
	PayerID string `json:"payer_id" binding:"required"`
}

type paymentEligibilityResponse struct {
	AppointmentID string `json:"appointment_id"`
	ExpertID      string `json:"expert_id"`
	AmountVND     int64  `json:"amount_vnd"`
	ExpiresAt     int64  `json:"expires_at"`
}

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

	status, err := appappointment.ParsePaymentResultStatus(req.Status)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment result status", err.Error())
		return
	}

	if err := h.usecase.HandlePaymentResult(appappointment.HandlePaymentResultCommand{
		AppointmentID: req.AppointmentID,
		Status:        status,
	}); err != nil {
		respondPaymentResultError(c, err)
		return
	}

	if status == appappointment.PaymentResultSuccess {
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

func (h *Handler) InternalPaymentEligibility(c *gin.Context) {
	if !isPaymentServiceCaller(c) {
		response.Error(c, http.StatusForbidden, "Forbidden", "caller is not allowed to read payment eligibility")
		return
	}

	appointmentID := strings.TrimSpace(c.Param("id"))
	if appointmentID == "" {
		response.Error(c, http.StatusBadRequest, "appointment_id is required", "missing path param :id")
		return
	}
	if _, err := uuid.Parse(appointmentID); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid appointment_id", err.Error())
		return
	}

	var req paymentEligibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment eligibility payload", err.Error())
		return
	}

	payerID := strings.TrimSpace(req.PayerID)
	if payerID == "" {
		response.Error(c, http.StatusBadRequest, "payer_id is required", "payer_id is required")
		return
	}
	if _, err := uuid.Parse(payerID); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payer_id", err.Error())
		return
	}

	eligibility, err := h.usecase.GetPaymentEligibility(appappointment.GetPaymentEligibilityCommand{
		AppointmentID: appointmentID,
		PayerID:       payerID,
	})
	if err != nil {
		respondPaymentEligibilityError(c, err)
		return
	}

	response.Success(c, "Payment eligibility verified", paymentEligibilityResponse{
		AppointmentID: eligibility.AppointmentID,
		ExpertID:      eligibility.ExpertID,
		AmountVND:     eligibility.AmountVND,
		ExpiresAt:     eligibility.ExpiresAt,
	})
}

func isPaymentServiceCaller(c *gin.Context) bool {
	callerID, ok := internal_auth.GetCallerID(c)
	return ok && callerID == "payment-service"
}

func respondPaymentEligibilityError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, appappointment.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Appointment not found", err.Error())
	case errors.Is(err, appappointment.ErrPaymentEligibilityForbidden):
		response.Error(c, http.StatusForbidden, "Appointment does not belong to payer", err.Error())
	case errors.Is(err, appappointment.ErrPaymentEligibilityConflict):
		response.Error(c, http.StatusConflict, "Appointment is not eligible for payment", err.Error())
	case errors.Is(err, appappointment.ErrInvalidBookingPrice):
		response.Error(c, http.StatusBadRequest, "Invalid booking price", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Payment eligibility check failed", err.Error())
	}
}

func respondPaymentResultError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, appappointment.ErrInvalidPaymentResultStatus):
		response.Error(c, http.StatusBadRequest, "Invalid payment result status", err.Error())
	case errors.Is(err, appappointment.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Appointment not found", err.Error())
	case errors.Is(err, appappointment.ErrInvalidStatus), errors.Is(err, appappointment.ErrPaymentResultConflict):
		response.Error(c, http.StatusConflict, "Payment result conflicts with booking state", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Payment result update failed", err.Error())
	}
}
