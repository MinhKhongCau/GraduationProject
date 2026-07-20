package handler

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"
	"booking-service/internal/slot"
	"booking-service/pkg/response"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetExpert handles GET /api/v1/booking/slots/expert.
// @Summary [EXPERT] Get expert slot calendar
// @Tags Slots
// @Security BearerAuth
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param status query string false "AVAILABLE, LOCKED, OCCUPIED, UNAVAILABLE"
// @Param availability_id query string false "Availability ID"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /booking/slots/expert [get]
func (h *Handler) GetExpert(c *gin.Context) {
	if c.GetHeader("X-User-Role") != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can access this endpoint", "Forbidden")
		return
	}
	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
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
	status, err := parseSlotStatus(c.Query("status"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid slot filter", err.Error())
		return
	}
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slots", "slot reader unavailable")
		return
	}
	result, err := h.reader.ListExpertSlots(slot.ExpertSlotQuery{ExpertID: expertID, FromMs: dateRange.FromMs, ToMs: dateRange.ToMs, Status: status, AvailabilityID: strings.TrimSpace(c.Query("availability_id")), Page: page})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve slots", err.Error())
		return
	}
	response.Success(c, "Get slots successfully", gin.H{"items": result.Items, "slots": result.Items, "page": result.Page, "size": result.Size, "total_items": result.TotalItems, "total": result.TotalItems, "total_pages": result.TotalPages, "has_next": result.HasNext, "has_previous": result.HasPrevious})
}

func parseSlotStatus(value string) (*domain.SlotStatus, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}
	values := map[string]domain.SlotStatus{"0": domain.SlotStatusAvailable, "AVAILABLE": domain.SlotStatusAvailable, "1": domain.SlotStatusLocked, "LOCKED": domain.SlotStatusLocked, "2": domain.SlotStatusOccupied, "OCCUPIED": domain.SlotStatusOccupied, "3": domain.SlotStatusUnavailable, "UNAVAILABLE": domain.SlotStatusUnavailable}
	status, ok := values[value]
	if !ok {
		return nil, fmt.Errorf("unsupported slot status")
	}
	return &status, nil
}
