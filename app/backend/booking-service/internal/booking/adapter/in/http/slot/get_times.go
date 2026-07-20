package handler

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/slot"
	"booking-service/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTimes handles GET /api/v1/public/booking/slots/available-times.
// @Summary [PUBLIC] Get bookable times for a date
// @Tags Slots
// @Param expert_id query string true "Expert ID"
// @Param date query string true "Date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /public/booking/slots/available-times [get]
func (h *Handler) GetTimes(c *gin.Context) {
	expertID, date := c.Query("expert_id"), c.Query("date")
	if expertID == "" || date == "" {
		response.Error(c, http.StatusBadRequest, "Missing 'expert_id' or 'date' parameter", "Missing parameters")
		return
	}
	if _, err := bookingquery.ParseDateRange(date, date, time.Now()); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid date", err.Error())
		return
	}
	page, err := bookingquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slot filter", err.Error())
		return
	}
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to query timeslots", "slot reader unavailable")
		return
	}
	result, err := h.reader.ListAvailableTimes(slot.AvailableTimeQuery{ExpertID: expertID, Date: date, Page: page})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to query timeslots", err.Error())
		return
	}
	response.Success(c, "Get available timeslots successfully", gin.H{"expert_id": expertID, "date": date, "items": result.Items, "available_times": result.Items, "page": result.Page, "size": result.Size, "total_items": result.TotalItems, "total_pages": result.TotalPages, "has_next": result.HasNext, "has_previous": result.HasPrevious})
}
