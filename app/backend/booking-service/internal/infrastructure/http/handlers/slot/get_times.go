package handler

import (
	bookingquery "booking-service/internal/application/query"
	"booking-service/internal/application/slot"
	"booking-service/internal/infrastructure/http/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTimes handles GET /api/v1/public/booking/slots/available-times.
// @Summary [PUBLIC] Get bookable times for a date
// @Tags Slots
// @Param expert_id query string true "Expert ID" default("2c230afb-a1af-4813-8b26-b17ae7fceb26") example("2c230afb-a1af-4813-8b26-b17ae7fceb26")
// @Param date query string true "Date YYYY-MM-DD" default("2026-07-25") example("2026-07-25")
// @Param page query int false "Zero-based page" default(0) example(0)
// @Param size query int false "Page size, 1-100" default(20) example(20)
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
