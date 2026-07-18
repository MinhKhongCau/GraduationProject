package handler

import (
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 1. POST /api/v1/booking/templates (Admin)
type CreateTemplateRequest struct {
	ShiftName           string `json:"shift_name" binding:"required"`
	StartTime           string `json:"start_time" binding:"required"` // "HH:MM"
	EndTime             string `json:"end_time" binding:"required"`   // "HH:MM"
	SlotDurationMinutes int    `json:"slot_duration_minutes" binding:"required,min=10,max=180"`
}

// CreateTemplate - POST /api/v1/booking/templates
//
//	@Summary      [ADMIN] Create a shift template
//	@Description  [ADMIN] Create a system-wide shift template (e.g., Morning Shift 08:00-12:00)
//	@Tags         Schedules
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateTemplateRequest  true  "Thông tin ca mẫu"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/templates [post]
func (h *Handler) CreateTemplate(c *gin.Context) {
	// Authorization: Only ADMIN can create templates
	userRole := c.GetHeader("X-User-Role")
	if userRole != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Only admin is authorized to create shift templates", "Forbidden")
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	template, err := h.usecase.CreateTimeTemplate(req.ShiftName, req.StartTime, req.EndTime, req.SlotDurationMinutes)
	if err != nil {
		writeScheduleError(c, err, "Failed to save shift template")
		return
	}

	response.Success(c, "Shift template created successfully!", template)
}

// GetTemplates - GET /api/v1/public/booking/templates
//
//	@Summary      [PUBLIC] Get all shift templates
//	@Description  [PUBLIC] Retrieve all system-wide shift templates
//	@Tags         Schedules
//	@Produce      json
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]interface{}
//	@Router       /public/booking/templates [get]
func (h *Handler) GetTemplates(c *gin.Context) {
	templates, err := h.usecase.GetTimeTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve shift templates", err.Error())
		return
	}
	response.Success(c, "Get shift templates successfully", templates)
}

type UpdateTemplateRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}

// UpdateTemplate - PATCH /api/v1/booking/templates/:id
//
//	@Summary      [ADMIN] Update shift template
//	@Description  [ADMIN] Update shift template (e.g. set is_active=false to deactivate)
//	@Tags         Schedules
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string  true  "Template ID"
//	@Param        body  body      UpdateTemplateRequest  true  "Thông tin cập nhật"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/templates/{id} [patch]
func (h *Handler) UpdateTemplate(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Only admin can update templates", "Forbidden")
		return
	}

	templateID := c.Param("id")

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.usecase.UpdateTemplate(templateID, *req.IsActive); err != nil {
		// Example of using sentinel error:
		// if errors.Is(err, schedule.ErrTemplateNotFound) {
		// 	response.Error(c, http.StatusNotFound, "Template not found", err.Error())
		// 	return
		// }
		writeScheduleError(c, err, "Failed to update template")
		return
	}

	response.Success(c, "Template updated successfully", gin.H{"id": templateID, "is_active": *req.IsActive})
}
