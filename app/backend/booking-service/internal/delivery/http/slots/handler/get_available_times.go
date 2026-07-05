package handler

import (
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetAvailableTimesHandler struct {
	repo *postgres.SlotRepository
}

func NewGetAvailableTimesHandler(repo *postgres.SlotRepository) *GetAvailableTimesHandler {
	return &GetAvailableTimesHandler{repo: repo}
}

// Handle - GET /api/v1/slots/available-times?date=2026-07-05&expert_id=xxx
// Trả về danh sách Slot với slot_id, start_time và end_time dạng Unix timestamp 13 số (ms)
//
//	@Summary      Lấy danh sách khung giờ trống trong ngày
//	@Description  Trả về danh sách slot khả dụng của một ngày, gồm slot_id và timestamp Unix 13 số
//	@Tags         Slots
//	@Produce      json
//	@Param        date       query     string  true  "Ngày cần tra cứu (YYYY-MM-DD)"
//	@Param        expert_id  query     string  true  "ID của chuyên gia"
//	@Success      200        {object}  map[string]interface{}
//	@Failure      400        {object}  map[string]interface{}
//	@Failure      500        {object}  map[string]interface{}
//	@Router       /slots/available-times [get]
func (h *GetAvailableTimesHandler) Handle(c *gin.Context) {
	dateParam := c.Query("date")
	if dateParam == "" {
		response.Error(c, http.StatusBadRequest, "Thiếu tham số 'date' (VD: ?date=2026-07-05)", "Missing date parameter")
		return
	}

	expertID := c.Query("expert_id")
	if expertID == "" {
		response.Error(c, http.StatusBadRequest, "Thiếu tham số 'expert_id'", "Missing expert_id parameter")
		return
	}

	slots, err := h.repo.GetAvailableTimes(dateParam, expertID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi truy vấn giờ", err.Error())
		return
	}

	response.Success(c, "Lấy danh sách khung giờ thành công", gin.H{
		"date":       dateParam,
		"expert_id":  expertID,
		"time_slots": slots, // [{slot_id, start_time (ms), end_time (ms)}, ...]
	})
}
