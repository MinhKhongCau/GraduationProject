package handler

import (
	"booking-service/internal/booking/application/appointment"
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SaveMedicalRecord handles POST / PUT /api/v1/booking/appointments/:id/medical-record.
// @Summary [EXPERT] Save or update medical record for an appointment
// @Tags Medical Records
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID"
// @Param payload body appointment.SaveMedicalRecordCommand true "Medical Record details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /booking/appointments/{id}/medical-record [post]
func (h *Handler) SaveMedicalRecord(c *gin.Context) {
	role := c.GetHeader("X-User-Role")
	if role != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can save medical records", "Forbidden")
		return
	}
	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	appointmentID := c.Param("id")
	if appointmentID == "" {
		response.Error(c, http.StatusBadRequest, "Appointment ID is required", "Missing appointment ID")
		return
	}

	var cmd appointment.SaveMedicalRecordCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid medical record data", err.Error())
		return
	}

	record, err := h.usecase.SaveMedicalRecord(expertID, role, appointmentID, cmd)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrUnauthorized):
			response.Error(c, http.StatusForbidden, "Access denied", err.Error())
		case errors.Is(err, appointment.ErrNotFound):
			response.Error(c, http.StatusNotFound, "Appointment not found", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to save medical record", err.Error())
		}
		return
	}

	response.Success(c, "Medical record saved successfully", record)
}

// GetAppointmentMedicalRecord handles GET /api/v1/booking/appointments/:id/medical-record.
// @Summary [PATIENT/EXPERT/ADMIN] Get medical record for an appointment
// @Tags Medical Records
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /booking/appointments/{id}/medical-record [get]
func (h *Handler) GetAppointmentMedicalRecord(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	role := c.GetHeader("X-User-Role")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	appointmentID := c.Param("id")
	record, err := h.usecase.GetMedicalRecordByAppointmentID(userID, role, appointmentID)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrUnauthorized):
			response.Error(c, http.StatusForbidden, "Access denied", err.Error())
		case errors.Is(err, appointment.ErrNotFound), errors.Is(err, appointment.ErrMedicalRecordNotFound):
			response.Error(c, http.StatusNotFound, "Medical record not found", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to get medical record", err.Error())
		}
		return
	}

	response.Success(c, "Get medical record successfully", record)
}

// ListMedicalRecords handles GET /api/v1/booking/medical-records.
// @Summary [PATIENT/EXPERT/ADMIN] List medical records of current user
// @Tags Medical Records
// @Produce json
// @Security BearerAuth
// @Param page query int false "Zero-based page index"
// @Param size query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /booking/medical-records [get]
func (h *Handler) ListMedicalRecords(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	role := c.GetHeader("X-User-Role")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	pageReq, err := bookingquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid pagination query", err.Error())
		return
	}

	page, err := h.usecase.ListMedicalRecords(userID, role, pageReq)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrUnauthorized):
			response.Error(c, http.StatusForbidden, "Access denied", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to list medical records", err.Error())
		}
		return
	}

	response.Success(c, "Get medical records successfully", gin.H{
		"items":         page.Items,
		"records":       page.Items,
		"page":          page.Page,
		"size":          page.Size,
		"total_items":   page.TotalItems,
		"total":         page.TotalItems,
		"total_pages":   page.TotalPages,
		"has_next":      page.HasNext,
		"has_previous":  page.HasPrevious,
	})
}

// GetMedicalRecordDetail handles GET /api/v1/booking/medical-records/:id.
// @Summary [PATIENT/EXPERT/ADMIN] Get medical record detail by record ID
// @Tags Medical Records
// @Produce json
// @Security BearerAuth
// @Param id path string true "Record ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /booking/medical-records/{id} [get]
func (h *Handler) GetMedicalRecordDetail(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	role := c.GetHeader("X-User-Role")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	recordID := c.Param("id")
	record, err := h.usecase.GetMedicalRecordByID(userID, role, recordID)
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrUnauthorized):
			response.Error(c, http.StatusForbidden, "Access denied", err.Error())
		case errors.Is(err, appointment.ErrNotFound), errors.Is(err, appointment.ErrMedicalRecordNotFound):
			response.Error(c, http.StatusNotFound, "Medical record not found", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to get medical record", err.Error())
		}
		return
	}

	response.Success(c, "Get medical record successfully", record)
}
