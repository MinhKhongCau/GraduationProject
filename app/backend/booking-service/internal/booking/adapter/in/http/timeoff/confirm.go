package handler

import (
	"booking-service/pkg/response"
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

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	timeOff, err := h.usecase.ConfirmTimeOff(expertID, req.StartDatetime, req.EndDatetime, req.Reason)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save time-off registration", err.Error())
		return
	}

	response.Success(c, "Time-off confirmed and registered successfully. Background worker will cancel overlapping appointments.", gin.H{
		"time_off": timeOff,
	})
}
