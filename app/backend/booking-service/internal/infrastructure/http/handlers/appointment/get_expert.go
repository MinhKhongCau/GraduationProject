package handler

import (
	"booking-service/internal/infrastructure/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetExpert handles GET /api/v1/booking/appointments/expert.
// @Summary [EXPERT] Get expert appointment calendar
// @Tags Appointments
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param status query string false "PENDING_PAYMENT, CONFIRMED, or CANCELLED"
// @Param patient_id query string false "Patient ID"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /booking/appointments/expert [get]
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
	filter, err := appointmentQuery(c, expertID, "EXPERT")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid appointment filter", err.Error())
		return
	}
	filter.PatientID = c.Query("patient_id")
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve appointments", "appointment reader unavailable")
		return
	}
	page, err := h.reader.ListAppointments(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve appointments", err.Error())
		return
	}
	response.Success(c, "Get appointments successfully", appointmentPageData(page))
}
