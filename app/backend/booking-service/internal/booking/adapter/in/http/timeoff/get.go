package handler

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/timeoff"
	"booking-service/pkg/response"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Get handles GET /api/v1/booking/time-off.
// @Summary [EXPERT] Get time-off calendar
// @Tags TimeOff
// @Security BearerAuth
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param processed query bool false "Processed state"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /booking/time-off [get]
func (h *Handler) Get(c *gin.Context) {
	if c.GetHeader("X-User-Role") != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can view their time-offs", "Forbidden")
		return
	}
	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Missing X-User-Id")
		return
	}
	page, err := bookingquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid time-off filter", err.Error())
		return
	}
	dateRange, err := bookingquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid time-off filter", err.Error())
		return
	}
	var processed *bool
	if value := strings.TrimSpace(c.Query("processed")); value != "" {
		parsed, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "Invalid time-off filter", "processed must be true or false")
			return
		}
		processed = &parsed
	}
	result, err := h.usecase.ListTimeOffs(timeoff.TimeOffListQuery{ExpertID: expertID, FromMs: dateRange.FromMs, ToMs: dateRange.ToMs, Processed: processed, Page: page})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve time-offs", err.Error())
		return
	}
	response.Success(c, "Get time-offs successfully", gin.H{"items": result.Items, "time_offs": result.Items, "page": result.Page, "size": result.Size, "total_items": result.TotalItems, "total": result.TotalItems, "total_pages": result.TotalPages, "has_next": result.HasNext, "has_previous": result.HasPrevious})
}
