package handler

import (
	"booking-service/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GenerateRequest struct {
	DaysToGenerate int `json:"days_to_generate" binding:"required,min=1,max=30"`
}

// Generate - POST /api/v1/slots/generate
// Phân quyền: Chỉ EXPERT mới được phép sinh lịch cho chính mình
//
//	@Summary      [EXPERT] Auto generate slots for expert
//	@Description  [EXPERT] MANUAL TRIGGER: Expert triggers slot generation immediately for next N days. (Note: System Cronjob is the primary source that runs automatically every night).
//	@Tags         Slots
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      GenerateRequest  true  "Expert ID và số ngày cần sinh"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/slots/generate [post]
func (h *Handler) Generate(c *gin.Context) {
	// ---- AUTHORIZATION: Only EXPERT can generate schedules ----
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts have permission to generate schedules", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "Missing expert ID", "Unauthorized")
		return
	}

	var req GenerateRequest

	// Validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Retrieve inputs from Database
	avails, err := h.scheduleRepo.GetAvailabilities(expertID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve availability config", err.Error())
		return
	}

	templates, err := h.scheduleRepo.GetTimeTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve time templates", err.Error())
		return
	}

	timeOffs, err := h.timeoffRepo.GetTimeOffs(expertID, time.Now())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve time-off configurations", err.Error())
		return
	}

	// Generate slots
	generatedSlots, err := h.usecase.GenerateSlotsForNextDays(expertID, req.DaysToGenerate, avails, templates, timeOffs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Slot generation algorithm error", err.Error())
		return
	}

	if len(generatedSlots) == 0 {
		response.Success(c, "No slots were generated (expert may be on leave or availability has not been configured)", nil)
		return
	}

	// Bulk Insert + ON CONFLICT DO NOTHING (Idempotent)
	if err := h.repo.BulkInsertSlots(generatedSlots); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save slots to database", err.Error())
		return
	}

	response.Success(c, "Slots generated successfully!", gin.H{
		"expert_id":     expertID,
		"slots_created": len(generatedSlots),
	})
}
