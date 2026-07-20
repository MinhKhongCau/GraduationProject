package handler

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/slot"
	"booking-service/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetDates handles GET /api/v1/public/booking/slots/available-dates.
// @Summary [PUBLIC] Get dates with bookable slots
// @Tags Slots
// @Param expert_id query string true "Expert ID"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /public/booking/slots/available-dates [get]
func (h *Handler) GetDates(c *gin.Context) {
	expertID := c.Query("expert_id")
	if expertID == "" {
		response.Error(c, http.StatusBadRequest, "Missing 'expert_id' parameter", "Missing expert_id parameter")
		return
	}
	page, err := bookingquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slot filter", err.Error())
		return
	}
	dateRange, err := bookingquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slot filter", err.Error())
		return
	}
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to query dates", "slot reader unavailable")
		return
	}
	result, err := h.reader.ListAvailableDates(slot.AvailableDateQuery{ExpertID: expertID, FromMs: dateRange.FromMs, ToMs: dateRange.ToMs, Page: page})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to query dates", err.Error())
		return
	}
	response.Success(c, "Get available dates successfully", gin.H{"expert_id": expertID, "items": result.Items, "available_dates": result.Items, "page": result.Page, "size": result.Size, "total_items": result.TotalItems, "total_pages": result.TotalPages, "has_next": result.HasNext, "has_previous": result.HasPrevious})
}
