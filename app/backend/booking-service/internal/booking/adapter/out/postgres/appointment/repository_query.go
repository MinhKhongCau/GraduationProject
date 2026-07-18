package appointmentpostgres

import "booking-service/internal/booking/domain"

// Láº¥y Appointment theo SlotID (Æ°u tiÃªn láº¥y cuá»™c háº¹n chÆ°a bá»‹ huá»·, náº¿u khÃ´ng thÃ¬ láº¥y cuá»™c háº¹n má»›i nháº¥t)
func (r *pgRepository) GetAppointmentBySlotID(slotID string) (*domain.Appointment, error) {
	var appt domain.Appointment
	err := r.db.Where("slot_id = ? AND status != ?", slotID, domain.AppointmentStatusCancelled).Order("created_at DESC").First(&appt).Error
	if err != nil {
		err = r.db.Where("slot_id = ?", slotID).Order("created_at DESC").First(&appt).Error
		if err != nil {
			return nil, err
		}
	}
	return &appt, nil
}

// Láº¥y Appointment theo AppointmentID
func (r *pgRepository) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	var appt domain.Appointment
	err := r.db.Table("Booking_Appointments").
		Select("\"Booking_Appointments\".*, \"Booking_Expert_Slots\".price").
		Joins("JOIN \"Booking_Expert_Slots\" ON \"Booking_Appointments\".slot_id = \"Booking_Expert_Slots\".slot_id").
		Where("\"Booking_Appointments\".appointment_id = ?", appointmentID).
		First(&appt).Error
	if err != nil {
		return nil, err
	}
	return &appt, nil
}

// Láº¥y danh sÃ¡ch cuá»™c háº¹n cá»§a Patient
func (r *pgRepository) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	var appts []domain.Appointment
	err := r.db.Where("patient_id = ?", patientID).Order("created_at desc").Find(&appts).Error
	return appts, err
}

// Láº¥y danh sÃ¡ch cuá»™c háº¹n cá»§a Expert
func (r *pgRepository) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	var appts []domain.Appointment
	query := r.db.Where("expert_id = ?", expertID)

	if fromDate > 0 {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate > 0 {
		query = query.Where("created_at <= ?", toDate)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at desc").Find(&appts).Error
	return appts, err
}
