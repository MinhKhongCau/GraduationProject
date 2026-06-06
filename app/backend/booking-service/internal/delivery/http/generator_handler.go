package http

import (
	"booking-service/internal/repository/postgres"
	"booking-service/internal/usecase"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	// 2. Lấy nguyên liệu từ Database (Repository)
	avails, err := h.repo.GetAvailabilities(req.ExpertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy lịch rảnh"})
		return
	}

	templates, err := h.repo.GetTimeTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy ca làm việc"})
		return
	}

	timeOffs, err := h.repo.GetTimeOffs(req.ExpertID, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy lịch nghỉ phép"})
		return
	}

	// 3. Đưa nguyên liệu vào máy xay (Usecase - Hàm Goroutines siêu tốc)
	generatedSlots, err := usecase.GenerateSlotsForNextDays(req.ExpertID, req.DaysToGenerate, avails, templates, timeOffs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi thuật toán cắt lịch: " + err.Error()})
		return
	}

	if len(generatedSlots) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Không có lịch nào được tạo (Có thể do bác sĩ nghỉ phép hoặc chưa cấu hình)"})
		return
	}

	// 4. Lưu toàn bộ sản phẩm xuống Database bằng lệnh Bulk Insert
	if err := h.repo.BulkInsertSlots(generatedSlots); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lưu lịch xuống DB: " + err.Error()})
		return
	}

	// 5. Trả kết quả thành công!
	c.JSON(http.StatusOK, gin.H{
		"message":       "Sinh lịch thành công rực rỡ!",
		"expert_id":     req.ExpertID,
		"slots_created": len(generatedSlots),
	})
}
