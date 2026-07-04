package http

import (
	"booking-service/internal/domain"
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AppointmentHandler xử lý các API liên quan đến việc đặt lịch khám
type AppointmentHandler struct {
	repo *postgres.AppointmentRepository
}

func NewAppointmentHandler(repo *postgres.AppointmentRepository) *AppointmentHandler {
	return &AppointmentHandler{repo: repo}
}

// =====================================================================
// 1. KHÓA SLOT TẠM THỜI
// POST /api/v1/slots/:id/lock
// Phân quyền: Chỉ PATIENT mới được giữ chỗ
// =====================================================================

// HandleLockSlot - POST /api/v1/slots/:id/lock
//
//	@Summary      Giữ chỗ slot tạm thời 15 phút
//	@Description  Bệnh nhân click chọn giờ, hệ thống khóa slot trong 15 phút để tiến hành thanh toán
//	@Tags         Appointments
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Slot ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]interface{}
//	@Failure      403  {object}  map[string]interface{}
//	@Failure      409  {object}  map[string]interface{}
//	@Router       /slots/{id}/lock [post]
func (h *AppointmentHandler) HandleLockSlot(c *gin.Context) {
	// Chỉ PATIENT mới được giữ chỗ
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Chỉ bệnh nhân mới có thể đặt lịch", "Forbidden")
		return
	}

	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", "Missing X-User-Id header")
		return
	}

	slotID := c.Param("id")
	if slotID == "" {
		response.Error(c, http.StatusBadRequest, "Thiếu Slot ID trong đường dẫn", "Missing slot_id")
		return
	}

	// Thực hiện Atomic Lock
	if err := h.repo.LockSlot(slotID, patientID); err != nil {
		response.Error(c, http.StatusConflict, err.Error(), "Slot unavailable")
		return
	}

	response.Success(c, "Giữ chỗ thành công! Bạn có 15 phút để hoàn tất thanh toán.", gin.H{
		"slot_id":              slotID,
		"lock_duration_seconds": 900,
	})
}

// =====================================================================
// 2. TẠO CUỘC HẸN
// POST /api/v1/appointments
// Phân quyền: Chỉ PATIENT (và phải là người đang giữ slot đó)
// =====================================================================

// CreateAppointmentRequest - Dữ liệu Patient gửi lên để tạo cuộc hẹn
type CreateAppointmentRequest struct {
	SlotID   string `json:"slot_id" binding:"required"`
	ExpertID string `json:"expert_id" binding:"required"`
}

// HandleCreateAppointment - POST /api/v1/appointments
//
//	@Summary      Tạo cuộc hẹn mới
//	@Description  Bệnh nhân xác nhận đặt lịch sau khi đã giữ chỗ thành công, trạng thái PENDING_PAYMENT
//	@Tags         Appointments
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateAppointmentRequest  true  "Thông tin cuộc hẹn"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /appointments [post]
func (h *AppointmentHandler) HandleCreateAppointment(c *gin.Context) {
	// Chỉ PATIENT mới được đặt lịch
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Chỉ bệnh nhân mới có thể đặt lịch", "Forbidden")
		return
	}

	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", "Missing X-User-Id header")
		return
	}

	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	appointment := &domain.Appointment{
		AppointmentID: uuid.New().String(),
		SlotID:        req.SlotID,
		PatientID:     patientID,
		ExpertID:      req.ExpertID,
		Status:        domain.AppointmentStatusPendingPayment,
	}

	if err := h.repo.CreateAppointment(appointment); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), "Create appointment failed")
		return
	}

	response.Success(c, "Tạo cuộc hẹn thành công! Vui lòng hoàn tất thanh toán trong 15 phút.", gin.H{
		"appointment_id": appointment.AppointmentID,
		"status":         appointment.Status,
		"slot_id":        appointment.SlotID,
	})
}

// =====================================================================
// 3. MOCK WEBHOOK XỬ LÝ KẾT QUẢ THANH TOÁN
// POST /api/v1/appointments/webhook
// Trong thực tế: Cổng thanh toán (VNPAY/MoMo) sẽ gọi endpoint này
// =====================================================================

// PaymentWebhookRequest - Dữ liệu cổng thanh toán gửi về
type PaymentWebhookRequest struct {
	AppointmentID string `json:"appointment_id" binding:"required"`
	Status        string `json:"status" binding:"required"` // "SUCCESS" hoặc "FAILED"
}

// HandlePaymentWebhook - POST /api/v1/appointments/webhook
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
func (h *AppointmentHandler) HandlePaymentWebhook(c *gin.Context) {
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
			"status":         domain.AppointmentStatusConfirmed.String(),
		})
	} else {
		// Thanh toán thất bại: Mở lock tức thì (Worker sẽ dọn định kỳ nếu miss)
		response.Success(c, "Đã nhận thông báo thanh toán thất bại. Slot sẽ được mở lại sau ít phút.", gin.H{
			"appointment_id": req.AppointmentID,
			"status":         domain.AppointmentStatusCancelled.String(),
		})
	}
}
