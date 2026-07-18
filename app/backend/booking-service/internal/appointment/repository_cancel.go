package appointment

import (
	"booking-service/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Há»§y Appointment do bÃ¡c sÄ© nghá»‰ phÃ©p (TimeOff)
func (r *pgRepository) CancelAppointmentByExpert(appointmentID string, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.Where("appointment_id = ?", appointmentID).First(&appt).Error; err != nil {
			return err
		}

		canceledBy := "EXPERT"
		if err := tx.Model(&appt).Updates(map[string]interface{}{
			"status":              domain.AppointmentStatusCancelled,
			"cancellation_reason": reason,
			"cancelled_by":        &canceledBy,
			"updated_at":          time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		// Tráº£ Slot vá» AVAILABLE
		if err := tx.Model(&domain.ExpertSlot{}).
			Where("slot_id = ?", appt.SlotID).
			Updates(map[string]interface{}{
				"status":            domain.SlotStatusAvailable,
				"locked_expires_at": nil,
				"locked_by":         nil,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// Há»§y Appointment do bá»‡nh nhÃ¢n (Patient)
func (r *pgRepository) CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var appt domain.Appointment
		if err := tx.Where("appointment_id = ? AND patient_id = ?", appointmentID, patientID).First(&appt).Error; err != nil {
			return errors.New("khÃ´ng tÃ¬m tháº¥y cuá»™c háº¹n hoáº·c báº¡n khÃ´ng cÃ³ quyá»n há»§y")
		}

		if appt.Status == domain.AppointmentStatusCancelled {
			return errors.New("cuá»™c háº¹n Ä‘Ã£ bá»‹ há»§y trÆ°á»›c Ä‘Ã³")
		}

		canceledBy := "PATIENT"
		if err := tx.Model(&appt).Updates(map[string]interface{}{
			"status":              domain.AppointmentStatusCancelled,
			"cancellation_reason": reason,
			"cancelled_by":        &canceledBy,
			"updated_at":          time.Now().UnixMilli(),
		}).Error; err != nil {
			return err
		}

		// Tráº£ Slot vá» AVAILABLE
		if err := tx.Model(&domain.ExpertSlot{}).
			Where("slot_id = ?", appt.SlotID).
			Updates(map[string]interface{}{
				"status":            domain.SlotStatusAvailable,
				"locked_expires_at": nil,
				"locked_by":         nil,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
