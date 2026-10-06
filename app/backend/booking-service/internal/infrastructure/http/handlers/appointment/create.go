package handler

import (
	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/infrastructure/http/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateRequest - Dữ liệu Patient gửi lên để tạo cuộc hẹn
type CreateRequest struct {
	SlotID           string `json:"slot_id" binding:"required"`
	ExpertID         string `json:"expert_id" binding:"required"`
	PatientRecordID  string `json:"patient_record_id" binding:"required"`
	SpecializationID string `json:"specialization_id"`
}

// Create - POST /api/v1/appointments
//
//	@Summary      [PATIENT] Create a new appointment
//	@Description  [PATIENT] Patient confirms booking after successfully locking a slot, state becomes PENDING_PAYMENT.
//	@Description  The chosen patient record is verified and snapshotted via profile-service gRPC.
//	@Tags         Appointments
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateRequest  true  "Thông vị cuộc hẹn"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      404   {object}  map[string]interface{}
//	@Failure      409   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Failure      503   {object}  map[string]interface{}
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

	appointment, err := h.usecase.CreateAppointment(c.Request.Context(), appappointment.CreateAppointmentCommand{
		PatientID:        patientID,
		ExpertID:         req.ExpertID,
		SlotID:           req.SlotID,
		PatientRecordID:  req.PatientRecordID,
		SpecializationID: req.SpecializationID,
	})
	if err != nil {
		if httpStatus, ok := bookingErrorStatus(err); ok {
			response.Error(c, httpStatus, err.Error(), "Create appointment failed")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error(), "Create appointment failed")
		return
	}

	response.Success(c, "Appointment created successfully! Please complete payment within 15 minutes.", gin.H{
		"appointment_id": appointment.AppointmentID,
		"status":         appointment.Status,
		"slot_id":        appointment.SlotID,
		"patient":        appointment.Patient,
	})
}

// bookingErrorStatus map lỗi đặt lịch (profile-service, giữ chỗ) sang HTTP status.
func bookingErrorStatus(err error) (int, bool) {
	switch {
	case errors.Is(err, appappointment.ErrBookingProfileNotFound):
		return http.StatusNotFound, true
	case errors.Is(err, appappointment.ErrBookingProfileInvalid):
		return http.StatusBadRequest, true
	case errors.Is(err, appappointment.ErrBookingSpecializationMismatch):
		return http.StatusUnprocessableEntity, true
	case errors.Is(err, appappointment.ErrBookingSlotNotHeld), errors.Is(err, appappointment.ErrSlotAlreadyBooked):
		return http.StatusConflict, true
	case errors.Is(err, appappointment.ErrProfileServiceUnavailable):
		return http.StatusServiceUnavailable, true
	default:
		return 0, false
	}
}
