package handler

import (
	"booking-service/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetDates - GET /api/v1/slots/available-dates?expert_id=xxx
// Trả về danh sách NGÀY có ít nhất 1 slot trống (AVAILABLE, chưa bị khóa, chưa qua)
//
//	@Summary      [PUBLIC] Get dates with available slots
//	@Description  [PUBLIC] Returns an array of dates (YYYY-MM-DD) that have available slots within next 1 month
//	@Tags         Slots
//	@Produce      json
//	@Param        expert_id  query     string  true  "ID của chuyên gia"
//	@Success      200        {object}  map[string]interface{}
//	@Failure      400        {object}  map[string]interface{}
//	@Failure      500        {object}  map[string]interface{}
//	@Router       /public/booking/slots/available-dates [get]
func (h *Handler) GetDates(c *gin.Context) {
	expertID := c.Query("expert_id")
	if expertID == "" {
		response.Error(c, http.StatusBadRequest, "Missing 'expert_id' parameter", "Missing expert_id parameter")
		return
	}

	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, 1, 0) // Scan next 1 month

	// TODO ỨNG DỤNG REDIS NHA: Cache kết quả đọc
	// Vì API này gọi rất nhiều từ phía Patient, nên cache kết quả (expert_id + month) vào Redis trong vài phút.
	dates, err := h.usecase.GetDates(expertID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to query dates", err.Error())
		return
	}

	response.Success(c, "Get available dates successfully", gin.H{
		"expert_id":       expertID,
		"available_dates": dates,
	})
}
