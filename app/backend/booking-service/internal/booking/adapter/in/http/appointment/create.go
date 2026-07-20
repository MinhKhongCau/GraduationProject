package handler

import (
	"booking-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateRequest - Dữ liệu Patient gửi lên để tạo cuộc hẹn
type CreateRequest struct {
	SlotID   string `json:"slot_id" binding:"required"`
	ExpertID string `json:"expert_id" binding:"required"`
}

// Create - POST /api/v1/appointments
//
//	@Summary      [PATIENT] Create a new appointment
//	@Description  [PATIENT] Patient confirms booking after successfully locking a slot, state becomes PENDING_PAYMENT
//	@Tags         Appointments
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateRequest  true  "Thông vị cuộc hẹn"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/appointments [post]
func (h *Handler) Create(c *gin.Context) {
	// Only PATIENT is allowed to book
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Only patients are allowed to book appointments", "Forbidden")
		return
	}

	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body data", err.Error())
		return
	}

	appointment, err := h.usecase.CreateAppointment(patientID, req.ExpertID, req.SlotID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), "Create appointment failed")
		return
	}

	response.Success(c, "Appointment created successfully! Please complete payment within 15 minutes.", gin.H{
		"appointment_id": appointment.AppointmentID,
		"status":         appointment.Status,
		"slot_id":        appointment.SlotID,
	})
}
