package handler

import (
	"booking-service/internal/application/appointment"
	bookingquery "booking-service/internal/application/query"
	"booking-service/internal/infrastructure/http/response"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func appointmentQuery(c *gin.Context, actorID, actorRole string) (appointment.AppointmentListQuery, error) {
	page, err := bookingquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		return appointment.AppointmentListQuery{}, err
	}
	status, err := appointment.ParseAppointmentStatus(c.Query("status"))
	if err != nil {
		return appointment.AppointmentListQuery{}, err
	}
	filter := appointment.AppointmentListQuery{ActorID: actorID, ActorRole: actorRole, Status: status, Page: page}
	legacyFrom, legacyTo := strings.TrimSpace(c.Query("from_date")), strings.TrimSpace(c.Query("to_date"))
	if c.Query("from") == "" && c.Query("to") == "" && (legacyFrom != "" || legacyTo != "") {
		if legacyFrom != "" {
			filter.FromMs, err = strconv.ParseInt(legacyFrom, 10, 64)
			if err != nil {
				return filter, bookingquery.ErrInvalidQuery
			}
		}
		if legacyTo != "" {
			filter.ToMs, err = strconv.ParseInt(legacyTo, 10, 64)
			if err != nil {
				return filter, bookingquery.ErrInvalidQuery
			}
			filter.ToMs++
		}
		if filter.FromMs > 0 && filter.ToMs > 0 && filter.FromMs >= filter.ToMs {
			return filter, bookingquery.ErrInvalidQuery
		}
		return filter, nil
	}
	rangeValue, err := bookingquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		return filter, err
	}
	filter.FromMs, filter.ToMs = rangeValue.FromMs, rangeValue.ToMs
	return filter, nil
}

// GetDetail handles GET /api/v1/booking/appointments/:id.
// @Summary [PATIENT/EXPERT/ADMIN] Get appointment detail with expert and patient profiles (admin: managed experts only)
// @Tags Appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /booking/appointments/{id} [get]
func (h *Handler) GetDetail(c *gin.Context) {
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve appointment", "appointment reader unavailable")
		return
	}
	result, err := h.reader.GetAppointmentDetail(c.Request.Context(), c.GetHeader("X-User-Id"), c.GetHeader("X-User-Role"), c.Param("id"))
	if err != nil {
		writeAppointmentReadError(c, err, "Failed to retrieve appointment")
		return
	}
	response.Success(c, "Get appointment successfully", result)
}

// GetAdmin handles GET /api/v1/booking/appointments/admin.
// @Summary [ADMIN] List appointments of experts I manage (approved by me)
// @Tags Appointments
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date YYYY-MM-DD (slot start time)"
// @Param to query string false "End date YYYY-MM-DD"
// @Param status query string false "PENDING_PAYMENT, CONFIRMED, CANCELLED, COMPLETED"
// @Param expert_id query string false "Expert auth ID (must be managed by me)"
// @Param patient_id query string false "Patient auth ID"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /booking/appointments/admin [get]
func (h *Handler) GetAdmin(c *gin.Context) {
	if c.GetHeader("X-User-Role") != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Only administrators can access this endpoint", "Forbidden")
		return
	}
	adminID := c.GetHeader("X-User-Id")
	if adminID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}
	filter, err := appointmentQuery(c, adminID, "ADMIN")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid appointment filter", err.Error())
		return
	}
	filter.ExpertID = c.Query("expert_id")
	filter.PatientID = c.Query("patient_id")
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve appointments", "appointment reader unavailable")
		return
	}
	page, err := h.reader.ListAdminAppointments(c.Request.Context(), filter)
	if err != nil {
		writeAppointmentReadError(c, err, "Failed to retrieve appointments")
		return
	}
	response.Success(c, "Get appointments successfully", appointmentPageData(page))
}

func writeAppointmentReadError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, appointment.ErrUnauthorized):
		response.Error(c, http.StatusForbidden, "Appointment access denied", err.Error())
	case errors.Is(err, appointment.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Appointment not found", err.Error())
	case errors.Is(err, appointment.ErrProfileServiceUnavailable):
		response.Error(c, http.StatusServiceUnavailable, "Cannot determine managed experts right now", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, fallback, err.Error())
	}
}
