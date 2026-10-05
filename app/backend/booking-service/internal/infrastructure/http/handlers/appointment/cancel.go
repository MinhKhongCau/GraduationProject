package handler

import (
	"booking-service/internal/infrastructure/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CancelRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// Cancel - PATCH /api/v1/appointments/:id/cancel
//
//	@Summary      [PATIENT/EXPERT] Cancel appointment
//	@Description  [PATIENT/EXPERT] Cancel an appointment and free up the slot. Patient can only cancel their own appointments.
//	@Tags         Appointments
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string                   true  "Appointment ID"
//	@Param        body  body      CancelRequest            true  "Lý do hủy"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      401   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      404   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/appointments/{id}/cancel [patch]
func (h *Handler) Cancel(c *gin.Context) {
	appointmentID := c.Param("id")
	if appointmentID == "" {
		response.Error(c, http.StatusBadRequest, "Missing Appointment ID in path", "Missing id")
		return
	}

	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	userRole := c.GetHeader("X-User-Role")
	userID := c.GetHeader("X-User-Id")

	if userID == "" || userRole == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing headers")
		return
	}

	err := h.usecase.CancelAppointment(appointmentID, userID, userRole, req.Reason)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to cancel appointment", err.Error())
		return
	}

	response.Success(c, "Appointment cancelled successfully", gin.H{
		"appointment_id": appointmentID,
	})
}
