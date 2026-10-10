package appointment

import (
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"fmt"
	"time"
)

// DefaultMinimumExpertPresence: chuyên gia phải có mặt trong phòng họp ít nhất 30 phút
// thì buổi tư vấn mới được tính là hoàn tất và được chi trả.
const DefaultMinimumExpertPresence = 30 * time.Minute

var (
	ErrSessionForbidden      = errors.New("only the appointment's expert can complete the session")
	ErrSessionNotCompletable = errors.New("appointment is not in a state that can complete a session")
	ErrSessionNotStarted     = errors.New("session has not started yet")
	ErrSessionTooShort       = errors.New("expert presence is shorter than the required minimum")
)

type SessionCompletionTransition struct {
	AppointmentUpdates map[string]interface{}
	// AlreadyCompleted: buổi tư vấn đã kết thúc trước đó (gọi lại để chi trả lại một cách idempotent).
	AlreadyCompleted bool
}

// PlanSessionCompletion kiểm tra điều kiện kết thúc buổi tư vấn:
//   - chỉ chuyên gia của lịch hẹn mới được kết thúc
//   - lịch hẹn đã thanh toán (CONFIRMED, hoặc COMPLETED do lưu bệnh án trước đó)
//   - đã tới giờ hẹn và chuyên gia có mặt đủ minimumPresence
func PlanSessionCompletion(appointment Appointment, slot slotdomain.ExpertSlot, expertID string, presence, minimumPresence time.Duration, nowMs int64) (*SessionCompletionTransition, error) {
	if appointment.ExpertID != expertID {
		return nil, ErrSessionForbidden
	}
	if appointment.SessionCompletedAt != nil {
		return &SessionCompletionTransition{AlreadyCompleted: true}, nil
	}
	if appointment.Status != AppointmentStatusConfirmed && appointment.Status != AppointmentStatusCompleted {
		return nil, fmt.Errorf("%w: status %s", ErrSessionNotCompletable, appointment.Status.String())
	}
	if nowMs < slot.StartTime {
		return nil, ErrSessionNotStarted
	}
	if presence < minimumPresence {
		return nil, fmt.Errorf("%w: %s < %s", ErrSessionTooShort, presence.Truncate(time.Second), minimumPresence)
	}
	return &SessionCompletionTransition{
		AppointmentUpdates: map[string]interface{}{
			"status":               AppointmentStatusCompleted,
			"session_completed_at": nowMs,
			"updated_at":           nowMs,
		},
	}, nil
}
