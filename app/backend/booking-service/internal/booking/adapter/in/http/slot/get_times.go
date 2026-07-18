package handler

import (
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTimes - GET /api/v1/slots/available-times?expert_id=xxx&date=YYYY-MM-DD
//
//	@Summary      [PUBLIC] Get available timeslots for a date
//	@Description  [PUBLIC] Returns an array of timeslots that are AVAILABLE for a specific expert on a specific date
//	@Tags         Slots
//	@Produce      json
//	@Param        expert_id  query     string  true  "ID của chuyên gia"
//	@Param        date       query     string  true  "Ngày cần xem lịch (YYYY-MM-DD)"
//	@Success      200        {object}  map[string]interface{}
//	@Failure      400        {object}  map[string]interface{}
//	@Failure      500        {object}  map[string]interface{}
//	@Router       /public/booking/slots/available-times [get]
func (h *Handler) GetTimes(c *gin.Context) {
	expertID := c.Query("expert_id")
	date := c.Query("date")

	if expertID == "" || date == "" {
		response.Error(c, http.StatusBadRequest, "Missing 'expert_id' or 'date' parameter", "Missing parameters")
		return
	}

	// TODO ỨNG DỤNG REDIS NHA: Cache kết quả đọc khung giờ của 1 chuyên gia trong 1 ngày
	times, err := h.usecase.GetTimes(expertID, date)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to query timeslots", err.Error())
		return
	}

	response.Success(c, "Get available timeslots successfully", gin.H{
		"expert_id":       expertID,
		"date":            date,
		"available_times": times,
	})
}
