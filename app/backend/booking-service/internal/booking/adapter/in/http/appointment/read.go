package handler

import (
	"booking-service/internal/booking/application/appointment"
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/pkg/response"
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
// @Summary [PATIENT/EXPERT/ADMIN] Get appointment detail
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
	result, err := h.reader.GetAppointmentDetail(c.GetHeader("X-User-Id"), c.GetHeader("X-User-Role"), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, appointment.ErrUnauthorized):
			response.Error(c, http.StatusForbidden, "Appointment access denied", err.Error())
		case errors.Is(err, appointment.ErrNotFound):
			response.Error(c, http.StatusNotFound, "Appointment not found", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve appointment", err.Error())
		}
		return
	}
	response.Success(c, "Get appointment successfully", result)
}
