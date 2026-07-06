package handler

import (
	"booking-service/internal/domain"
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateTimeOffHandler struct {
	timeoffRepo     *postgres.TimeOffRepository
	slotRepo        *postgres.SlotRepository
	appointmentRepo *postgres.AppointmentRepository
}

func NewCreateTimeOffHandler(
	timeoffRepo *postgres.TimeOffRepository,
	slotRepo *postgres.SlotRepository,
	appointmentRepo *postgres.AppointmentRepository,
) *CreateTimeOffHandler {
	return &CreateTimeOffHandler{
		timeoffRepo:     timeoffRepo,
		slotRepo:        slotRepo,
		appointmentRepo: appointmentRepo,
	}
}

type CreateTimeOffRequest struct {
	StartDatetime int64  `json:"start_datetime" binding:"required"`
	EndDatetime   int64  `json:"end_datetime" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
	Force         bool   `json:"force"`
}

func (h *CreateTimeOffHandler) Handle(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Chỉ chuyên gia mới có quyền đăng ký nghỉ phép", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính người dùng", "Missing X-User-Id header")
		return
	}

	var req CreateTimeOffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	// 1. Kiểm tra va chạm (Collision Detection)
	overlappingSlots, err := h.timeoffRepo.GetOverlappingSlots(expertID, req.StartDatetime, req.EndDatetime)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi kiểm tra lịch đụng", err.Error())
		return
	}

	hasOccupied := false
	var affectedAppointments []string
	
	for _, slot := range overlappingSlots {
		if slot.Status == domain.SlotStatusOccupied {
			hasOccupied = true
			// Tìm appointment_id bị ảnh hưởng
			appt, _ := h.appointmentRepo.GetAppointmentBySlotID(slot.SlotID)
			if appt != nil {
				affectedAppointments = append(affectedAppointments, appt.AppointmentID)
			}
		}
	}

	// 2. Nếu có slot đang OCCUPIED và không có cờ force=true -> Cảnh báo
	if hasOccupied && !req.Force {
		affectedAppointmentsStr := strings.Join(affectedAppointments, ", ")
		response.Error(c, http.StatusConflict, 
			"Khoảng thời gian này đã có lịch hẹn xác nhận. Gửi lại với force=true để ép buộc nghỉ (sẽ tự động huỷ lịch bệnh nhân).", 
			"affected_appointments: "+affectedAppointmentsStr)
		return
	}

	// 3. Cho phép lưu TimeOff
	timeOff := domain.ExpertTimeOff{
		TimeOffID:     uuid.New().String(),
		ExpertID:      expertID,
		StartDatetime: req.StartDatetime,
		EndDatetime:   req.EndDatetime,
		Reason:        req.Reason,
		// ProcessedAt để trống (null) để cho Background Worker xử lý huỷ lịch
	}

	if err := h.timeoffRepo.CreateTimeOff(&timeOff); err != nil {
		response.Error(c, http.StatusInternalServerError, "Lưu lịch nghỉ thất bại", err.Error())
		return
	}

	response.Success(c, "Đã lưu lịch nghỉ thành công. Hệ thống đang tự động dọn dẹp các lịch ảnh hưởng.", gin.H{
		"time_off": timeOff,
	})
}
