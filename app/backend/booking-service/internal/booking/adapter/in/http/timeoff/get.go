package handler

import (
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get - GET /api/v1/booking/time-off
//
//	@Summary      [EXPERT] Get list of time-offs
//	@Description  [EXPERT] Retrieve all time-off registrations for the current expert
//	@Tags         TimeOff
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/time-off [get]
func (h *Handler) Get(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can view their time-offs", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")

	timeOffs, err := h.usecase.GetTimeOffs(expertID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve time-offs", err.Error())
		return
	}

	response.Success(c, "Get time-offs successfully", gin.H{
		"time_offs": timeOffs,
		"total":     len(timeOffs),
	})
}
