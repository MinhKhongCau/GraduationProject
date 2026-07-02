package http

import (
	"booking-service/internal/repository/postgres"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SlotHandler struct {
	repo *postgres.SlotRepository
}

func NewSlotHandler(repo *postgres.SlotRepository) *SlotHandler {
	return &SlotHandler{repo: repo}
}

// API 1: /api/v1/slots/available-dates?month=04&year=2026
func (h *SlotHandler) HandleGetAvailableDates(c *gin.Context) {
	// Lấy tháng và năm từ URL (Frontend gửi lên)
	// Để đơn giản test ngay, nếu không truyền ta sẽ quét 30 ngày từ hôm nay
	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, 1, 0) // Cộng thêm 1 tháng

	dates, err := h.repo.GetAvailableDates(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi truy vấn ngày: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Lấy danh sách ngày thành công",
		"available_dates": dates,
	})
}

// API 2: /api/v1/slots/available-times?date=2026-04-21
func (h *SlotHandler) HandleGetAvailableTimes(c *gin.Context) {
	// Bắt buộc phải có ngày truyền lên
	dateParam := c.Query("date")
	if dateParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tham số 'date' (VD: ?date=2026-04-21)"})
		return
	}

	times, err := h.repo.GetAvailableTimes(dateParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi truy vấn giờ: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Lấy danh sách giờ thành công",
		"date":       dateParam,
		"time_slots": times,
	})
}
