package appointment

import (
	"booking-service/internal/domain"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =====================================================================
// 3. XÃC NHáº¬N THANH TOÃN THÃ€NH CÃ”NG (Webhook xá»­ lÃ½ sau khi cá»•ng TT gá»i vÃ o)
// Cáº­p nháº­t appointment -> CONFIRMED, slot -> OCCUPIED vÃ  giáº£i phÃ³ng lock
// =====================================================================
func (r *pgRepository) ConfirmPayment(appointmentID string) error {
	return r.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultSuccess,
	})
}

// HandlePaymentResult applies payment SUCCESS/FAILED state transitions atomically.
func (r *pgRepository) HandlePaymentResult(command HandlePaymentResultCommand) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("appointment_id = ?", command.AppointmentID).
			First(&appt).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		var slot domain.ExpertSlot
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("slot_id = ?", appt.SlotID).
			First(&slot).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: slot not found", ErrPaymentResultConflict)
			}
			return err
		}

		nowMs := time.Now().UnixMilli()
		switch command.Status {
		case PaymentResultSuccess:
			return r.applyPaymentSuccess(tx, &appt, &slot, nowMs)
		case PaymentResultFailed:
			return r.applyPaymentFailure(tx, &appt, &slot, nowMs)
		default:
			return ErrInvalidPaymentResultStatus
		}
	})
}

func (r *pgRepository) applyPaymentSuccess(tx *gorm.DB, appt *domain.Appointment, slot *domain.ExpertSlot, nowMs int64) error {
	transition, err := planPaymentResultTransition(*appt, *slot, PaymentResultSuccess, nowMs)
	if err != nil {
		return err
	}
	return r.persistPaymentResultTransition(tx, appt, transition)
}

func (r *pgRepository) applyPaymentFailure(tx *gorm.DB, appt *domain.Appointment, slot *domain.ExpertSlot, nowMs int64) error {
	transition, err := planPaymentResultTransition(*appt, *slot, PaymentResultFailed, nowMs)
	if err != nil {
		return err
	}
	return r.persistPaymentResultTransition(tx, appt, transition)
}

type paymentResultTransition struct {
	appointmentUpdates map[string]interface{}
	slotUpdates        map[string]interface{}
	expectedSlotStatus domain.SlotStatus
	noop               bool
}

func planPaymentResultTransition(appt domain.Appointment, slot domain.ExpertSlot, status PaymentResultStatus, nowMs int64) (*paymentResultTransition, error) {
	switch status {
	case PaymentResultSuccess:
		return planPaymentSuccess(appt, slot, nowMs)
	case PaymentResultFailed:
		return planPaymentFailure(appt, slot, nowMs)
	default:
		return nil, ErrInvalidPaymentResultStatus
	}
}

func planPaymentSuccess(appt domain.Appointment, slot domain.ExpertSlot, nowMs int64) (*paymentResultTransition, error) {
	switch appt.Status {
	case domain.AppointmentStatusPendingPayment:
		if slot.Status != domain.SlotStatusLocked {
			return nil, fmt.Errorf("%w: success requires LOCKED slot, got %s", ErrPaymentResultConflict, slot.Status.String())
		}
		confirmedAt := nowMs
		return &paymentResultTransition{
			appointmentUpdates: map[string]interface{}{
				"status":       domain.AppointmentStatusConfirmed,
				"updated_at":   nowMs,
				"confirmed_at": confirmedAt,
			},
			slotUpdates: map[string]interface{}{
				"status":            domain.SlotStatusOccupied,
				"locked_expires_at": nil,
				"locked_by":         nil,
			},
			expectedSlotStatus: domain.SlotStatusLocked,
		}, nil
	case domain.AppointmentStatusConfirmed:
		return &paymentResultTransition{noop: true}, nil
	case domain.AppointmentStatusCancelled:
		return nil, fmt.Errorf("%w: cancelled appointment cannot be confirmed", ErrPaymentResultConflict)
	default:
		return nil, fmt.Errorf("%w: unsupported appointment status %d", ErrPaymentResultConflict, appt.Status)
	}
}

func planPaymentFailure(appt domain.Appointment, slot domain.ExpertSlot, nowMs int64) (*paymentResultTransition, error) {
	switch appt.Status {
	case domain.AppointmentStatusPendingPayment:
		if slot.Status != domain.SlotStatusLocked {
			return nil, fmt.Errorf("%w: failure requires LOCKED slot, got %s", ErrPaymentResultConflict, slot.Status.String())
		}
		cancelledBy := "PAYMENT"
		return &paymentResultTransition{
			appointmentUpdates: map[string]interface{}{
				"status":              domain.AppointmentStatusCancelled,
				"cancellation_reason": "Payment failed",
				"cancelled_by":        &cancelledBy,
				"updated_at":          nowMs,
			},
			slotUpdates: map[string]interface{}{
				"status":            domain.SlotStatusAvailable,
				"locked_expires_at": nil,
				"locked_by":         nil,
			},
			expectedSlotStatus: domain.SlotStatusLocked,
		}, nil
	case domain.AppointmentStatusCancelled:
		return &paymentResultTransition{noop: true}, nil
	case domain.AppointmentStatusConfirmed:
		return nil, fmt.Errorf("%w: confirmed appointment cannot be cancelled by payment failure", ErrPaymentResultConflict)
	default:
		return nil, fmt.Errorf("%w: unsupported appointment status %d", ErrPaymentResultConflict, appt.Status)
	}
}

func (r *pgRepository) persistPaymentResultTransition(tx *gorm.DB, appt *domain.Appointment, transition *paymentResultTransition) error {
	if transition.noop {
		return nil
	}
	if err := tx.Model(appt).Updates(transition.appointmentUpdates).Error; err != nil {
		return err
	}

	result := tx.Model(&domain.ExpertSlot{}).
		Where("slot_id = ? AND status = ?", appt.SlotID, transition.expectedSlotStatus).
		Updates(transition.slotUpdates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: slot is no longer %s", ErrPaymentResultConflict, transition.expectedSlotStatus.String())
	}
	return nil
}
