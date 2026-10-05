package appointment

import (
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"fmt"
)

type PaymentResultStatus string

const (
	PaymentResultSuccess PaymentResultStatus = "SUCCESS"
	PaymentResultFailed  PaymentResultStatus = "FAILED"
)

var (
	ErrInvalidPaymentResultStatus = errors.New("invalid payment result status")
	ErrPaymentResultConflict      = errors.New("payment result conflicts with current appointment state")
)

type PaymentResultTransition struct {
	AppointmentUpdates map[string]interface{}
	SlotUpdates        map[string]interface{}
	ExpectedSlotStatus slotdomain.SlotStatus
	Noop               bool
}

func PlanPaymentResultTransition(appointment Appointment, slot slotdomain.ExpertSlot, status PaymentResultStatus, nowMs int64) (*PaymentResultTransition, error) {
	switch status {
	case PaymentResultSuccess:
		return planPaymentSuccess(appointment, slot, nowMs)
	case PaymentResultFailed:
		return planPaymentFailure(appointment, slot, nowMs)
	default:
		return nil, ErrInvalidPaymentResultStatus
	}
}

func planPaymentSuccess(appointment Appointment, slot slotdomain.ExpertSlot, nowMs int64) (*PaymentResultTransition, error) {
	switch appointment.Status {
	case AppointmentStatusPendingPayment:
		if slot.Status != slotdomain.SlotStatusLocked {
			return nil, fmt.Errorf("%w: success requires LOCKED slot, got %s", ErrPaymentResultConflict, slot.Status.String())
		}
		confirmedAt := nowMs
		return &PaymentResultTransition{
			AppointmentUpdates: map[string]interface{}{
				"status":       AppointmentStatusConfirmed,
				"updated_at":   nowMs,
				"confirmed_at": confirmedAt,
			},
			SlotUpdates: map[string]interface{}{
				"status":            slotdomain.SlotStatusOccupied,
				"locked_expires_at": nil,
				"locked_by":         nil,
			},
			ExpectedSlotStatus: slotdomain.SlotStatusLocked,
		}, nil
	case AppointmentStatusConfirmed:
		return &PaymentResultTransition{Noop: true}, nil
	case AppointmentStatusCancelled:
		return nil, fmt.Errorf("%w: cancelled appointment cannot be confirmed", ErrPaymentResultConflict)
	default:
		return nil, fmt.Errorf("%w: unsupported appointment status %d", ErrPaymentResultConflict, appointment.Status)
	}
}

func planPaymentFailure(appointment Appointment, slot slotdomain.ExpertSlot, nowMs int64) (*PaymentResultTransition, error) {
	switch appointment.Status {
	case AppointmentStatusPendingPayment:
		if slot.Status != slotdomain.SlotStatusLocked {
			return nil, fmt.Errorf("%w: failure requires LOCKED slot, got %s", ErrPaymentResultConflict, slot.Status.String())
		}
		cancelledBy := "PAYMENT"
		return &PaymentResultTransition{
			AppointmentUpdates: map[string]interface{}{
				"status":              AppointmentStatusCancelled,
				"cancellation_reason": "Payment failed",
				"cancelled_by":        &cancelledBy,
				"updated_at":          nowMs,
			},
			SlotUpdates: map[string]interface{}{
				"status":            slotdomain.SlotStatusAvailable,
				"locked_expires_at": nil,
				"locked_by":         nil,
			},
			ExpectedSlotStatus: slotdomain.SlotStatusLocked,
		}, nil
	case AppointmentStatusCancelled:
		return &PaymentResultTransition{Noop: true}, nil
	case AppointmentStatusConfirmed:
		return nil, fmt.Errorf("%w: confirmed appointment cannot be cancelled by payment failure", ErrPaymentResultConflict)
	default:
		return nil, fmt.Errorf("%w: unsupported appointment status %d", ErrPaymentResultConflict, appointment.Status)
	}
}
