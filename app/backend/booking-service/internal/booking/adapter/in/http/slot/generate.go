package handler

import (
	"booking-service/internal/slot"
	"booking-service/pkg/response"
	"errors"
	"net/http"

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

	result, err := h.generation.GenerateExpert(c.Request.Context(), expertID, req.DaysToGenerate)
	if err != nil {
		switch {
		case errors.Is(err, slot.ErrInvalidGeneration):
			response.Error(c, http.StatusBadRequest, "Invalid slot generation configuration", err.Error())
		case errors.Is(err, slot.ErrSlotOverlap):
			response.Error(c, http.StatusConflict, "Expert slot schedule overlaps", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Slot generation failed", "Internal server error")
		}
		return
	}

	if result.Candidates == 0 {
		response.Success(c, "No slots were generated (expert may be on leave or availability has not been configured)", nil)
		return
	}

	response.Success(c, "Slots generated successfully!", gin.H{
		"expert_id":     expertID,
		"slots_created": result.Inserted,
	})
}
