package handler

import (
	"booking-service/internal/booking/domain"
	"booking-service/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetExpert - GET /api/v1/appointments/expert
//
//	@Summary      [EXPERT] Get expert's appointments
//	@Description  [EXPERT] Get a list of appointments for the current expert with optional filters
//	@Tags         Appointments
//	@Produce      json
//	@Security     BearerAuth
//	@Param        from_date  query     int     false  "From Date (Unix ms)"
//	@Param        to_date    query     int     false  "To Date (Unix ms)"
//	@Param        status     query     int     false  "Status (0=PENDING, 1=CONFIRMED, 2=CANCELLED)"
//	@Success      200   {object}  map[string]interface{}
//	@Failure      401   {object}  map[string]interface{}
//	@Failure      403   {object}  map[string]interface{}
//	@Failure      500   {object}  map[string]interface{}
//	@Router       /booking/appointments/expert [get]
func (h *Handler) GetExpert(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can access this endpoint", "Forbidden")
		return
	}

	expertID := c.GetHeader("X-User-Id")
	if expertID == "" {
		response.Error(c, http.StatusUnauthorized, "User identity could not be determined", "Missing X-User-Id header")
		return
	}

	// Parse query params
	var fromDate, toDate int64
	var status *domain.AppointmentStatus

	if fd := c.Query("from_date"); fd != "" {
		if val, err := strconv.ParseInt(fd, 10, 64); err == nil {
			fromDate = val
		}
	}
	if td := c.Query("to_date"); td != "" {
		if val, err := strconv.ParseInt(td, 10, 64); err == nil {
			toDate = val
		}
	}
	if st := c.Query("status"); st != "" {
		if val, err := strconv.Atoi(st); err == nil {
			s := domain.AppointmentStatus(val)
			status = &s
		}
	}

	appts, err := h.usecase.GetAppointmentsByExpert(expertID, fromDate, toDate, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve appointments", err.Error())
		return
	}

	response.Success(c, "Get appointments successfully", gin.H{
		"appointments": appts,
		"total":        len(appts),
	})
}
