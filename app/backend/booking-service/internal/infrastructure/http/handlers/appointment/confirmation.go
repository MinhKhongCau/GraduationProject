package handler

import (
	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/infrastructure/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ConfirmationQuery - Tham số trang xác nhận đặt lịch
type ConfirmationQuery struct {
	SlotID           string `form:"slot_id" binding:"required"`
	ExpertID         string `form:"expert_id" binding:"required"`
	PatientRecordID  string `form:"patient_record_id" binding:"required"`
	SpecializationID string `form:"specialization_id"`
}

// GetConfirmation - GET /api/v1/booking/appointments/confirmation
//
//	@Summary      [PATIENT] Booking confirmation details
//	@Description  [PATIENT] After locking a slot, returns slot, expert, specialization and patient record
//	@Description  (queried from profile-service via gRPC) for the confirm-booking page.
//	@Tags         Appointments
//	@Produce      json
//	@Security     BearerAuth
//	@Param        slot_id            query     string  true   "Locked slot ID"
//	@Param        expert_id          query     string  true   "Expert account ID"
//	@Param        patient_record_id  query     string  true   "Patient record ID"
//	@Param        specialization_id  query     string  false  "Selected specialization ID"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      400   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      404   {object}  map[string]interface{}
//	@Failure      409   {object}  map[string]interface{}
//	@Failure      503   {object}  map[string]interface{}
//	@Router       /booking/appointments/confirmation [get]
func (h *Handler) GetConfirmation(c *gin.Context) {
	if c.GetHeader("X-User-Role") != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Only patients are allowed to book appointments", "Forbidden")
		return
	}
	patientID := c.GetHeader("X-User-Id")
	if patientID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	var query ConfirmationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	confirmer, ok := h.usecase.(appappointment.BookingConfirmationUsecase)
	if !ok {
		response.Error(c, http.StatusInternalServerError, appappointment.ErrBookingConfirmationUnavailable.Error(), "Get booking confirmation failed")
		return
	}

	confirmation, err := confirmer.GetBookingConfirmation(c.Request.Context(), appappointment.BookingConfirmationQuery{
		PatientID:        patientID,
		ExpertID:         query.ExpertID,
		SlotID:           query.SlotID,
		PatientRecordID:  query.PatientRecordID,
		SpecializationID: query.SpecializationID,
	})
	if err != nil {
		if httpStatus, ok := bookingErrorStatus(err); ok {
			response.Error(c, httpStatus, err.Error(), "Get booking confirmation failed")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error(), "Get booking confirmation failed")
		return
	}

	response.Success(c, "Get booking confirmation successfully", confirmation)
}
