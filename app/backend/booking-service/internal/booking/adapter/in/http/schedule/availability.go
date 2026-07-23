package handler

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/schedule"
	"booking-service/pkg/response"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateAvailabilityRequest struct {
	TemplateID     string  `json:"template_id" binding:"required"`
	DayOfWeek      int     `json:"day_of_week" binding:"required,min=1,max=7"` // 1=Mon...7=Sun
	EffectiveFrom  int64   `json:"effective_from" binding:"required"`          // Unix ms
	EffectiveUntil *int64  `json:"effective_until"`
	Price          float64 `json:"price" binding:"required"`
}

// CreateAvailability - POST /api/v1/booking/availabilities
//
//	@Summary      [EXPERT] Create expert availability configuration
//	@Description  [EXPERT] Create weekly availability configuration for an expert based on templates
//	@Tags         Schedules
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateAvailabilityRequest  true  "Thông tin lịch rảnh"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/availabilities [post]
func (h *Handler) CreateAvailability(c *gin.Context) {
	// Authorization: Only EXPERT
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can register availability configurations", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "Identity could not be determined", "Missing X-User-Id")
		return
	}

	var req CreateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	avail, err := h.usecase.CreateAvailability(expertID, req.TemplateID, req.DayOfWeek, req.EffectiveFrom, req.EffectiveUntil, req.Price)
	if err != nil {
		writeScheduleError(c, err, "Failed to save availability configuration")
		return
	}

	response.Success(c, "Availability configuration registered successfully!", avail)
}

// GetAvailabilities - GET /api/v1/booking/availabilities
//
//	@Summary      [EXPERT] Get expert availability configurations
//	@Description  [EXPERT] Retrieve configurations of weekly availability for the logged-in expert
//	@Tags         Schedules
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {object}  map[string]interface{}
//	@Failure      403  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]interface{}
//	@Router       /booking/availabilities [get]
func (h *Handler) GetAvailabilities(c *gin.Context) {
	// Only EXPERT retrieves their own availability
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Forbidden", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Missing X-User-Id")
		return
	}

	page, err := bookingquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid availability filter", err.Error())
		return
	}
	var active *bool
	if value := strings.TrimSpace(c.Query("active")); value != "" {
		parsed, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "Invalid availability filter", "active must be true or false")
			return
		}
		active = &parsed
	}
	var dateRange bookingquery.DateRange
	if c.Query("effective_from") != "" || c.Query("effective_to") != "" {
		dateRange, err = bookingquery.ParseDateRange(c.Query("effective_from"), c.Query("effective_to"), time.Now())
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid availability filter", err.Error())
			return
		}
	}
	result, err := h.usecase.ListAvailabilities(schedule.AvailabilityListQuery{ExpertID: expertID, Active: active, EffectiveFromMs: dateRange.FromMs, EffectiveToMs: dateRange.ToMs, Page: page})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve availability configurations", err.Error())
		return
	}

	response.Success(c, "Get availability configurations successfully", gin.H{"items": result.Items, "availabilities": result.Items, "page": result.Page, "size": result.Size, "total_items": result.TotalItems, "total_pages": result.TotalPages, "has_next": result.HasNext, "has_previous": result.HasPrevious})
}

type UpdateAvailabilityRequest struct {
	TemplateID     *string  `json:"template_id"`
	DayOfWeek      *int     `json:"day_of_week"`
	IsEnabled      *bool    `json:"is_enabled"`
	EffectiveFrom  *int64   `json:"effective_from"`
	EffectiveUntil *int64   `json:"effective_until"`
	Price          *float64 `json:"price"`
}

// UpdateAvailability - PATCH /api/v1/booking/availabilities/:id
//
//	@Summary      [EXPERT] Update expert availability
//	@Description  [EXPERT] Update a specific availability configuration
//	@Tags         Schedules
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string  true  "Availability ID"
//	@Param        body  body      UpdateAvailabilityRequest  true  "Cập nhật (truyền các field cần thiết)"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/availabilities/{id} [patch]
func (h *Handler) UpdateAvailability(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Forbidden", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	availID := c.Param("id")

	var req UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.TemplateID != nil {
		updates["template_id"] = *req.TemplateID
	}
	if req.DayOfWeek != nil {
		updates["day_of_week"] = *req.DayOfWeek
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.EffectiveFrom != nil {
		updates["effective_from"] = *req.EffectiveFrom
	}
	if req.EffectiveUntil != nil {
		updates["effective_until"] = *req.EffectiveUntil
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, "No updates provided", "Empty payload")
		return
	}

	if err := h.usecase.UpdateAvailability(availID, expertID, updates); err != nil {
		writeScheduleError(c, err, "Failed to update availability")
		return
	}

	response.Success(c, "Availability updated successfully", gin.H{"id": availID, "updates": updates})
}

func writeScheduleError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, schedule.ErrInvalidSchedule):
		response.Error(c, http.StatusBadRequest, "Invalid schedule configuration", err.Error())
	case errors.Is(err, schedule.ErrScheduleOverlap):
		response.Error(c, http.StatusConflict, "Schedule configuration overlaps", err.Error())
	case errors.Is(err, schedule.ErrScheduleNotFound):
		response.Error(c, http.StatusNotFound, "Schedule configuration not found", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, fallback, "Internal server error")
	}
}
