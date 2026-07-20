package handler

import (
	"booking-service/internal/timeoff"
	"booking-service/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Confirm - POST /api/v1/booking/time-off/confirm
//
//	@Summary      [EXPERT] Confirm and force register time-off
//	@Description  [EXPERT] Force saves the time-off and automatically cancels all overlapping appointments
//	@Tags         TimeOff
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateRequest  true  "Thông tin thời gian nghỉ phép"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/time-off/confirm [post]
func (h *Handler) Confirm(c *gin.Context) {
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

	timeOff, err := h.usecase.ConfirmTimeOff(expertID, req.StartDatetime, req.EndDatetime, req.Reason)
	if err != nil {
		if errors.Is(err, timeoff.ErrConflict) {
			response.Error(c, http.StatusConflict, "Time-off conflicts with a confirmed or occupied booking", err.Error())
			return
		}
		if errors.Is(err, timeoff.ErrInvalidDate) || errors.Is(err, timeoff.ErrDuplicate) {
			response.Error(c, http.StatusBadRequest, "Invalid time-off registration", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to save time-off registration", err.Error())
		return
	}

	response.Success(c, "Time-off confirmed and affected pending-payment bookings reconciled successfully.", gin.H{
		"time_off": timeOff,
	})
}
