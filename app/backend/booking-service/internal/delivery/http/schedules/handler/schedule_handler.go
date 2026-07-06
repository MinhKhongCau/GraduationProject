package handler

import (
	"booking-service/internal/domain"
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	repo *postgres.GeneratorRepository
}

func NewScheduleHandler(repo *postgres.GeneratorRepository) *ScheduleHandler {
	return &ScheduleHandler{repo: repo}
}

// 1. POST /api/v1/booking/templates (Admin)
type CreateTemplateRequest struct {
	ShiftName           string `json:"shift_name" binding:"required"`
	StartTime           string `json:"start_time" binding:"required"` // "HH:MM"
	EndTime             string `json:"end_time" binding:"required"`   // "HH:MM"
	SlotDurationMinutes int    `json:"slot_duration_minutes" binding:"required,min=10,max=180"`
}

func (h *ScheduleHandler) CreateTemplate(c *gin.Context) {
	// Phân quyền: Chỉ ADMIN được tạo template
	userRole := c.GetHeader("X-User-Role")
	if userRole != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Chỉ Admin mới có quyền tạo ca mẫu", "Forbidden")
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	template := domain.TimeTemplate{
		TemplateID:          uuid.New().String(),
		ShiftName:           req.ShiftName,
		StartTime:           req.StartTime,
		EndTime:             req.EndTime,
		SlotDurationMinutes: req.SlotDurationMinutes,
		IsActive:            true,
	}

	if err := h.repo.CreateTimeTemplate(&template); err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu ca mẫu", err.Error())
		return
	}

	response.Success(c, "Tạo ca mẫu thành công!", template)
}

// 2. GET /api/v1/public/booking/templates (Public)
func (h *ScheduleHandler) GetTemplates(c *gin.Context) {
	templates, err := h.repo.GetTimeTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lấy ca mẫu", err.Error())
		return
	}
	response.Success(c, "Lấy danh sách ca mẫu thành công", templates)
}

// 3. POST /api/v1/booking/availabilities (Expert)
type CreateAvailabilityRequest struct {
	TemplateID     string `json:"template_id" binding:"required"`
	DayOfWeek      int    `json:"day_of_week" binding:"required,min=1,max=7"` // 1=Mon...7=Sun
	EffectiveFrom  int64  `json:"effective_from" binding:"required"`          // Unix ms
	EffectiveUntil *int64 `json:"effective_until"`
}

func (h *ScheduleHandler) CreateAvailability(c *gin.Context) {
	// Phân quyền: Chỉ EXPERT
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Chỉ chuyên gia mới được đăng ký lịch rảnh", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định danh tính", "Missing X-User-Id")
		return
	}

	var req CreateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	avail := domain.Availability{
		AvailabilityID: uuid.New().String(),
		ExpertID:       expertID,
		TemplateID:     req.TemplateID,
		DayOfWeek:      req.DayOfWeek,
		IsEnabled:      true,
		EffectiveFrom:  req.EffectiveFrom,
		EffectiveUntil: req.EffectiveUntil,
	}

	if err := h.repo.CreateAvailability(&avail); err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lưu cấu hình lịch", err.Error())
		return
	}

	response.Success(c, "Đăng ký lịch rảnh thành công!", avail)
}

// 4. GET /api/v1/booking/availabilities (Expert)
func (h *ScheduleHandler) GetAvailabilities(c *gin.Context) {
	// Chỉ EXPERT lấy của chính mình
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Forbidden", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Missing X-User-Id")
		return
	}

	avails, err := h.repo.GetAvailabilities(expertID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lấy lịch rảnh", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách lịch rảnh thành công", avails)
}
