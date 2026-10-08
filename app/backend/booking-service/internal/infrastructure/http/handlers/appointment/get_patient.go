package handler

import (
	bookingquery "booking-service/internal/application/query"
	appointmentdomain "booking-service/internal/domain/appointment"
	"booking-service/internal/infrastructure/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPatient handles GET /api/v1/booking/appointments.
// @Summary [PATIENT] Get patient appointment calendar
// @Tags Appointments
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param status query string false "PENDING_PAYMENT, CONFIRMED, or CANCELLED"
// @Param expert_id query string false "Expert ID"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /booking/appointments [get]
func (h *Handler) GetPatient(c *gin.Context) {
	if c.GetHeader("X-User-Role") != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Only patients can access their appointments", "Forbidden")
		return
	}
	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}
	filter, err := appointmentQuery(c, patientID, "PATIENT")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid appointment filter", err.Error())
		return
	}
	filter.ExpertID = c.Query("expert_id")
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

func appointmentPageData(page bookingquery.Page[appointmentdomain.Appointment]) gin.H {
	return gin.H{"items": page.Items, "appointments": page.Items, "page": page.Page, "size": page.Size, "total_items": page.TotalItems, "total": page.TotalItems, "total_pages": page.TotalPages, "has_next": page.HasNext, "has_previous": page.HasPrevious}
}
