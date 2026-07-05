package handler

import (
	"booking-service/internal/domain"
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateAppointmentRequest - Dữ liệu Patient gửi lên để tạo cuộc hẹn
type CreateAppointmentRequest struct {
	SlotID   string `json:"slot_id" binding:"required"`
	ExpertID string `json:"expert_id" binding:"required"`
}

type CreateAppointmentHandler struct {
	repo *postgres.AppointmentRepository
}

func NewCreateAppointmentHandler(repo *postgres.AppointmentRepository) *CreateAppointmentHandler {
	return &CreateAppointmentHandler{repo: repo}
}

// Handle - POST /api/v1/appointments
//
//	@Summary      Tạo cuộc hẹn mới
//	@Description  Bệnh nhân xác nhận đặt lịch sau khi đã giữ chỗ thành công, trạng thái PENDING_PAYMENT
//	@Tags         Appointments
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateAppointmentRequest  true  "Thông vị cuộc hẹn"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /appointments [post]
func (h *CreateAppointmentHandler) Handle(c *gin.Context) {
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
