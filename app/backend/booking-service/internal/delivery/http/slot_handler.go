package http

import (
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/response"
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

// HandleGetAvailableDates - GET /api/v1/slots/available-dates?expert_id=xxx
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
func (h *SlotHandler) HandleGetAvailableDates(c *gin.Context) {
	expertID := c.Query("expert_id")
	if expertID == "" {
		response.Error(c, http.StatusBadRequest, "Thiếu tham số 'expert_id'", "Missing expert_id parameter")
		return
	}

	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, 1, 0) // Quét trong vòng 1 tháng tới

	// TODO: Nếu sau này muốn lọc theo expert thì thêm điều kiện expert_id vào GetAvailableDates
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

// HandleGetAvailableTimes - GET /api/v1/slots/available-times?date=2026-07-05&expert_id=xxx
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
func (h *SlotHandler) HandleGetAvailableTimes(c *gin.Context) {
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
