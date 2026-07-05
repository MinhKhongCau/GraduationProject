package handler

import (
	"booking-service/internal/repository/postgres"
	"booking-service/internal/usecase"
	"booking-service/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Khai báo cấu trúc nhận dữ liệu từ người dùng (Frontend gửi lên)
type GenerateRequest struct {
	ExpertID       string `json:"expert_id" binding:"required"`
	DaysToGenerate int    `json:"days_to_generate" binding:"required,min=1,max=30"`
}

type GenerateSlotsHandler struct {
	repo *postgres.GeneratorRepository
}

func NewGenerateSlotsHandler(repo *postgres.GeneratorRepository) *GenerateSlotsHandler {
	return &GenerateSlotsHandler{repo: repo}
}

// Handle - POST /api/v1/slots/generate
// Phân quyền: Chỉ EXPERT mới được phép sinh lịch cho chính mình
//
//	@Summary      Sinh lịch khám tự động cho chuyên gia
//	@Description  Expert tự kích hoạt để hệ thống sinh slot trong N ngày tới dựa trên cấu hình lịch rảnh
//	@Tags         Slots
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      GenerateRequest  true  "Expert ID và số ngày cần sinh"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /slots/generate [post]
func (h *GenerateSlotsHandler) Handle(c *gin.Context) {
	// ---- PHÂN QUYỀN: Chỉ EXPERT được sinh lịch ----
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Chỉ chuyên gia mới có quyền sinh lịch", "Forbidden")
		return
	}

	callerID := c.GetHeader("X-User-Id")

	var req GenerateRequest

	// Kiểm tra dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	// ---- PHÂN QUYỀN: Expert chỉ được sinh lịch cho chính mình ----
	if callerID != req.ExpertID {
		response.Error(c, http.StatusForbidden, "Bạn không thể sinh lịch thay cho chuyên gia khác", "Forbidden")
		return
	}

	// Lấy nguyên liệu từ Database (Repository)
	avails, err := h.repo.GetAvailabilities(req.ExpertID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi lấy lịch rảnh", err.Error())
		return
	}

	templates, err := h.repo.GetTimeTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi lấy ca làm việc", err.Error())
		return
	}

	timeOffs, err := h.repo.GetTimeOffs(req.ExpertID, time.Now())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi lấy lịch nghỉ phép", err.Error())
		return
	}

	// Đưa nguyên liệu vào máy xay (Usecase - Goroutines sinh lịch song song)
	generatedSlots, err := usecase.GenerateSlotsForNextDays(req.ExpertID, req.DaysToGenerate, avails, templates, timeOffs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi thuật toán cắt lịch", err.Error())
		return
	}

	if len(generatedSlots) == 0 {
		response.Success(c, "Không có lịch nào được tạo (Có thể do bác sĩ nghỉ phép hoặc chưa cấu hình)", nil)
		return
	}

	// Lưu xuống Database bằng Bulk Insert + ON CONFLICT DO NOTHING (Idempotent)
	if err := h.repo.BulkInsertSlots(generatedSlots); err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi lưu lịch xuống DB", err.Error())
		return
	}

	response.Success(c, "Sinh lịch thành công!", gin.H{
		"expert_id":     req.ExpertID,
		"slots_created": len(generatedSlots),
	})
}
