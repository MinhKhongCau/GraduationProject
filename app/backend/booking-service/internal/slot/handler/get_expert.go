package handler

import (
	"booking-service/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetExpert - GET /api/v1/slots/expert
//
//	@Summary      [EXPERT] Get all slots for expert
//	@Description  [EXPERT] Retrieve all slots (AVAILABLE, LOCKED, OCCUPIED) to display on calendar
//	@Tags         Slots
//	@Produce      json
//	@Security     BearerAuth
//	@Param        from_date  query     int     false  "From Date (Unix ms)"
//	@Param        to_date    query     int     false  "To Date (Unix ms)"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      401   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/slots/expert [get]
func (h *Handler) GetExpert(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can access this endpoint", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	var fromDate, toDate int64

	if fd := c.Query("from_date"); fd != "" {
		if val, err := strconv.ParseInt(fd, 10, 64); err == nil {
			fromDate = val
		}
	}
	if td := c.Query("to_date"); td != "" {
		if val, err := strconv.ParseInt(td, 10, 64); err == nil {
			toDate = val
		}
	}

	slots, err := h.repo.GetSlotsByExpert(expertID, fromDate, toDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slots", err.Error())
		return
	}

	response.Success(c, "Get slots successfully", gin.H{
		"slots": slots,
		"total": len(slots),
	})
}
