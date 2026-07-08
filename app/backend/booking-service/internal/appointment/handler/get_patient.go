package handler

import (
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPatient - GET /api/v1/appointments
//
//	@Summary      [PATIENT] Get patient's appointments
//	@Description  [PATIENT] Get a list of all appointments for the current patient
//	@Tags         Appointments
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200   {object}  map[string]interface{}
//	@Failure      401   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/appointments [get]
func (h *Handler) GetPatient(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Only patients can access their appointments", "Forbidden")
		return
	}

	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	appts, err := h.usecase.GetAppointmentsByPatient(patientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve appointments", err.Error())
		return
	}

	response.Success(c, "Get appointments successfully", gin.H{
		"appointments": appts,
		"total":        len(appts),
	})
}
