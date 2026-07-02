package http

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

// GeneratorHandler chứa các API liên quan đến việc tạo lịch
type GeneratorHandler struct {
	repo *postgres.GeneratorRepository
}

func NewGeneratorHandler(repo *postgres.GeneratorRepository) *GeneratorHandler {
	return &GeneratorHandler{repo: repo}
}

// HandleGenerateSlots là hàm xử lý khi có request POST gọi tới
func (h *GeneratorHandler) HandleGenerateSlots(c *gin.Context) {
	var req GenerateRequest

	// 1. Kiểm tra dữ liệu đầu vào xem có đúng chuẩn không
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}

	// 2. Lấy nguyên liệu từ Database (Repository)
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

	// 3. Đưa nguyên liệu vào máy xay (Usecase - Hàm Goroutines siêu tốc)
	generatedSlots, err := usecase.GenerateSlotsForNextDays(req.ExpertID, req.DaysToGenerate, avails, templates, timeOffs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi thuật toán cắt lịch", err.Error())
		return
	}

	if len(generatedSlots) == 0 {
		response.Success(c, "Không có lịch nào được tạo (Có thể do bác sĩ nghỉ phép hoặc chưa cấu hình)", nil)
		return
	}

	// 4. Lưu toàn bộ sản phẩm xuống Database bằng lệnh Bulk Insert
	if err := h.repo.BulkInsertSlots(generatedSlots); err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi lưu lịch xuống DB", err.Error())
		return
	}

	// 5. Trả kết quả thành công!
	response.Success(c, "Sinh lịch thành công rực rỡ!", gin.H{
		"expert_id":     req.ExpertID,
		"slots_created": len(generatedSlots),
	})
}
