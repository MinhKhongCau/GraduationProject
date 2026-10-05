package appointment

import (
	slotdomain "booking-service/internal/domain/slot"
	"errors"
)

type CancellationActor string

const (
	CancellationActorPatient CancellationActor = "PATIENT"
	CancellationActorExpert  CancellationActor = "EXPERT"
	CancellationActorSystem  CancellationActor = "SYSTEM"
)

var (
	ErrAppointmentAlreadyCancelled = errors.New("appointment already cancelled")
	ErrInvalidCancellationActor    = errors.New("invalid cancellation actor")
)

type CancellationPlan struct {
	AppointmentUpdates map[string]interface{}
	SlotUpdates        map[string]interface{}
	ShouldReleaseSlot  bool
}

func PlanAppointmentCancellation(appointment Appointment, actor CancellationActor, reason string, nowMs int64) (CancellationPlan, error) {
	switch actor {
	case CancellationActorPatient:
		if appointment.Status == AppointmentStatusCancelled {
			return CancellationPlan{}, ErrAppointmentAlreadyCancelled
		}
		return newCancellationPlan(actor, reason, nowMs), nil
	case CancellationActorExpert:
		return newCancellationPlan(actor, reason, nowMs), nil
	case CancellationActorSystem:
		return newCancellationPlan(actor, reason, nowMs), nil
	default:
		return CancellationPlan{}, ErrInvalidCancellationActor
	}
}

func newCancellationPlan(actor CancellationActor, reason string, nowMs int64) CancellationPlan {
	cancelledBy := string(actor)
	return CancellationPlan{
		AppointmentUpdates: map[string]interface{}{
			"status":              AppointmentStatusCancelled,
			"cancellation_reason": reason,
			"cancelled_by":        &cancelledBy,
			"updated_at":          nowMs,
		},
		SlotUpdates:       ReleasedSlotUpdates(),
		ShouldReleaseSlot: true,
	}
}

func ReleasedSlotUpdates() map[string]interface{} {
	return map[string]interface{}{
		"status":            slotdomain.SlotStatusAvailable,
		"locked_expires_at": nil,
		"locked_by":         nil,
	}
}
