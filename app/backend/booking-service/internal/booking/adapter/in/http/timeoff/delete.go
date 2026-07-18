package handler

import (
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Delete - DELETE /api/v1/booking/time-off/:id
//
//	@Summary      [EXPERT] Delete time-off
//	@Description  [EXPERT] Cancel a time-off registration. Note: cancelled appointments will NOT be automatically restored.
//	@Tags         TimeOff
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string  true  "TimeOff ID"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/time-off/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can delete their time-offs", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	timeOffID := c.Param("id")

	if err := h.usecase.DeleteTimeOff(expertID, timeOffID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete time-off", err.Error())
		return
	}

	response.Success(c, "Time-off deleted successfully", gin.H{
		"time_off_id": timeOffID,
	})
}
