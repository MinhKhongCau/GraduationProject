package repository

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *pgRepository) GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error) {
	var snapshot PaymentEligibilitySnapshot

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var appt appointmentdomain.Appointment
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("appointment_id = ?", command.AppointmentID).
			First(&appt).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		var slot slotdomain.ExpertSlot
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("slot_id = ?", appt.SlotID).
			First(&slot).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: slot not found", ErrPaymentEligibilityConflict)
			}
			return err
		}

		snapshot = PaymentEligibilitySnapshot{
			Appointment: appt,
			Slot:        slot,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
