package handler

import (
	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/infrastructure/http/middleware"
	"booking-service/internal/infrastructure/http/response"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MeetingCallerID là service duy nhất quản lý phòng họp (chatroom-service).
const MeetingCallerID = "chatroom-service"

type completeSessionRequest struct {
	ExpertID        string `json:"expert_id" binding:"required"`
	PresenceSeconds int64  `json:"presence_seconds"`
}

// InternalGetMeeting — GET /internal/meetings/:token
// chatroom-service gọi khi bệnh nhân/chuyên gia mở link phòng họp.
func (h *Handler) InternalGetMeeting(c *gin.Context) {
	meetings, ok := h.meetingUsecase(c)
	if !ok {
		return
	}
	info, err := meetings.GetMeetingByToken(c.Request.Context(), c.Param("token"))
	if err != nil {
		respondMeetingError(c, err)
		return
	}
	response.Success(c, "Get meeting successfully", info)
}

// InternalCompleteSession — POST /internal/appointments/:id/complete-session
// chatroom-service gọi khi chuyên gia kết thúc buổi tư vấn sau khi đã có mặt đủ thời gian.
func (h *Handler) InternalCompleteSession(c *gin.Context) {
	meetings, ok := h.meetingUsecase(c)
	if !ok {
		return
	}
	appointmentID := strings.TrimSpace(c.Param("id"))
	if _, err := uuid.Parse(appointmentID); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid appointment_id", err.Error())
		return
	}
	var req completeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid complete-session payload", err.Error())
		return
	}
	if req.PresenceSeconds < 0 {
		response.Error(c, http.StatusBadRequest, "presence_seconds must not be negative", "invalid_presence")
		return
	}

	result, err := meetings.CompleteSession(c.Request.Context(), appappointment.CompleteSessionCommand{
		AppointmentID: appointmentID,
		ExpertID:      strings.TrimSpace(req.ExpertID),
		Presence:      time.Duration(req.PresenceSeconds) * time.Second,
	})
	if err != nil {
		respondMeetingError(c, err)
		return
	}
	response.Success(c, "Session completed", result)
}

func (h *Handler) meetingUsecase(c *gin.Context) (appappointment.MeetingUsecase, bool) {
	if callerID, ok := middleware.GetCallerID(c); !ok || callerID != MeetingCallerID {
		response.Error(c, http.StatusForbidden, "Forbidden", "caller is not allowed to manage meetings")
		return nil, false
	}
	meetings, ok := h.usecase.(appappointment.MeetingUsecase)
	if !ok {
		response.Error(c, http.StatusNotImplemented, "Meetings are not configured", "not_implemented")
		return nil, false
	}
	return meetings, true
}

func respondMeetingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, appappointment.ErrMeetingNotFound), errors.Is(err, appappointment.ErrNotFound):
		response.Error(c, http.StatusNotFound, "Meeting not found", err.Error())
	case errors.Is(err, appappointment.ErrSessionForbidden):
		response.Error(c, http.StatusForbidden, "Only the appointment's expert can complete the session", err.Error())
	case errors.Is(err, appappointment.ErrSessionNotCompletable), errors.Is(err, appappointment.ErrSessionNotStarted), errors.Is(err, appappointment.ErrSessionTooShort):
		response.Error(c, http.StatusConflict, "Session cannot be completed yet", err.Error())
	case errors.Is(err, appappointment.ErrSessionSettlementFails):
		// Buổi tư vấn đã được ghi nhận; chatroom-service gọi lại để thử chi trả lần nữa.
		response.Error(c, http.StatusBadGateway, "Session completed but payout failed, retry later", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "Meeting operation failed", err.Error())
	}
}
