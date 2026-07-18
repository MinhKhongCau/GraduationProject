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
	transition, err := domain.PlanPaymentResultTransition(*appt, *slot, domain.PaymentResultSuccess, nowMs)
	if err != nil {
		return mapPaymentResultDomainError(err)
	}
	return r.persistPaymentResultTransition(tx, appt, transition)
}

func (r *pgRepository) applyPaymentFailure(tx *gorm.DB, appt *domain.Appointment, slot *domain.ExpertSlot, nowMs int64) error {
	transition, err := domain.PlanPaymentResultTransition(*appt, *slot, domain.PaymentResultFailed, nowMs)
	if err != nil {
		return mapPaymentResultDomainError(err)
	}
	return r.persistPaymentResultTransition(tx, appt, transition)
}

func mapPaymentResultDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidPaymentResultStatus):
		return fmt.Errorf("%w: %v", ErrInvalidPaymentResultStatus, err)
	case errors.Is(err, domain.ErrPaymentResultConflict):
		return fmt.Errorf("%w: %v", ErrPaymentResultConflict, err)
	default:
		return err
	}
}

func (r *pgRepository) persistPaymentResultTransition(tx *gorm.DB, appt *domain.Appointment, transition *domain.PaymentResultTransition) error {
	if transition.Noop {
		return nil
	}
	if err := tx.Model(appt).Updates(transition.AppointmentUpdates).Error; err != nil {
		return err
	}

	result := tx.Model(&domain.ExpertSlot{}).
		Where("slot_id = ? AND status = ?", appt.SlotID, transition.ExpectedSlotStatus).
		Updates(transition.SlotUpdates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: slot is no longer %s", ErrPaymentResultConflict, transition.ExpectedSlotStatus.String())
	}
	return nil
}
