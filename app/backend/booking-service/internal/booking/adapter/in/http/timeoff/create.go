package handler

import (
	"booking-service/internal/timeoff"
	"booking-service/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CreateRequest struct {
	StartDatetime int64  `json:"start_datetime" binding:"required"`
	EndDatetime   int64  `json:"end_datetime" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
}

// Create - POST /api/v1/booking/time-off
//
//	@Summary      [EXPERT] Register expert time-off
//	@Description  [EXPERT] Expert registers time-off period, returns 409 conflict if there are overlapping appointments
//	@Tags         TimeOff
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateRequest  true  "Thông tin thời gian nghỉ phép"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      409   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/time-off [post]
func (h *Handler) Create(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts have permission to register time-off", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	timeOff, affectedAppointments, err := h.usecase.CreateTimeOff(expertID, req.StartDatetime, req.EndDatetime, req.Reason)
	
	if err != nil {
		if err == timeoff.ErrConflict {
			affectedAppointmentsStr := strings.Join(affectedAppointments, ", ")
			response.Error(c, http.StatusConflict,
				"This time period already has confirmed appointments. Please call /time-off/confirm to override and automatically cancel affected appointments.",
				"affected_appointments: "+affectedAppointmentsStr)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to save time-off registration", err.Error())
		return
	}

	response.Success(c, "Time-off registered successfully. Affected schedules are being cleaned up.", gin.H{
		"time_off": timeOff,
	})
}
