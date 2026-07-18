package appointmentpostgres

import (
	"booking-service/internal/domain"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *pgRepository) GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error) {
	var snapshot PaymentEligibilitySnapshot

	err := r.db.Transaction(func(tx *gorm.DB) error {
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
