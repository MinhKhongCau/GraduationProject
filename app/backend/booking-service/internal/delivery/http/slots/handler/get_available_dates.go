package handler

import (
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GetAvailableDatesHandler struct {
	repo *postgres.SlotRepository
}

func NewGetAvailableDatesHandler(repo *postgres.SlotRepository) *GetAvailableDatesHandler {
	return &GetAvailableDatesHandler{repo: repo}
}

// Handle - GET /api/v1/slots/available-dates?expert_id=xxx
// Trả về danh sách NGÀY có ít nhất 1 slot trống (AVAILABLE, chưa bị khóa, chưa qua)
//
//	@Summary      Lấy danh sách ngày có lịch trống
//	@Description  Trả về mảng ngày định dạng "YYYY-MM-DD" có slot khả dụng trong 1 tháng tới
//	@Tags         Slots
//	@Produce      json
//	@Param        expert_id  query     string  true  "ID của chuyên gia"
//	@Success      200        {object}  map[string]interface{}
//	@Failure      400        {object}  map[string]interface{}
//	@Failure      500        {object}  map[string]interface{}
//	@Router       /slots/available-dates [get]
func (h *GetAvailableDatesHandler) Handle(c *gin.Context) {
	expertID := c.Query("expert_id")
	if expertID == "" {
		response.Error(c, http.StatusBadRequest, "Thiếu tham số 'expert_id'", "Missing expert_id parameter")
		return
	}

	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, 1, 0) // Quét trong vòng 1 tháng tới

	dates, err := h.repo.GetAvailableDates(startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi truy vấn ngày", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách ngày thành công", gin.H{
		"expert_id":       expertID,
		"available_dates": dates,
	})
}
