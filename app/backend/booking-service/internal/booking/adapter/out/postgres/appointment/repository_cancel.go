package appointmentpostgres

import (
	"booking-service/internal/booking/domain"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Há»§y Appointment do bÃ¡c sÄ© nghá»‰ phÃ©p (TimeOff)
func (r *pgRepository) CancelAppointmentByExpert(appointmentID string, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("appointment_id = ?", appointmentID).First(&appt).Error; err != nil {
			return err
		}

		plan, err := domain.PlanAppointmentCancellation(appt, domain.CancellationActorExpert, reason, time.Now().UnixMilli())
		if err != nil {
			return err
		}

		if err := tx.Model(&appt).Updates(plan.AppointmentUpdates).Error; err != nil {
			return err
		}

		// Tráº£ Slot vá» AVAILABLE
		if plan.ShouldReleaseSlot {
			var slot domain.ExpertSlot
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("slot_id = ?", appt.SlotID).First(&slot).Error; err != nil {
				return err
			}
			updates, err := releasedSlotUpdatesForCoverage(tx, slot)
			if err != nil {
				return err
			}
			if err := tx.Model(&slot).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Há»§y Appointment do bá»‡nh nhÃ¢n (Patient)
func (r *pgRepository) CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("appointment_id = ? AND patient_id = ?", appointmentID, patientID).First(&appt).Error; err != nil {
			return errors.New("khÃ´ng tÃ¬m tháº¥y cuá»™c háº¹n hoáº·c báº¡n khÃ´ng cÃ³ quyá»n há»§y")
		}

		plan, err := domain.PlanAppointmentCancellation(appt, domain.CancellationActorPatient, reason, time.Now().UnixMilli())
		if err != nil {
			return mapCancellationDomainError(err)
		}

		if err := tx.Model(&appt).Updates(plan.AppointmentUpdates).Error; err != nil {
			return err
		}

		// Tráº£ Slot vá» AVAILABLE
		if plan.ShouldReleaseSlot {
			var slot domain.ExpertSlot
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("slot_id = ?", appt.SlotID).First(&slot).Error; err != nil {
				return err
			}
			updates, err := releasedSlotUpdatesForCoverage(tx, slot)
			if err != nil {
				return err
			}
			if err := tx.Model(&slot).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func mapCancellationDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrAppointmentAlreadyCancelled):
		return errors.New("cuá»™c háº¹n Ä‘Ã£ bá»‹ há»§y trÆ°á»›c Ä‘Ã³")
	default:
		return err
	}
}
