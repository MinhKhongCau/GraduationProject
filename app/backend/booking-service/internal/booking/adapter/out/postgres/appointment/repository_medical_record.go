package appointmentpostgres

import (
	"booking-service/internal/booking/domain"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *pgRepository) UpdateAppointmentStatus(appointmentID string, status domain.AppointmentStatus) error {
	return r.db.Model(&domain.Appointment{}).Where("appointment_id = ?", appointmentID).Update("status", status).Error
}

func (r *pgRepository) SaveMedicalRecord(record *domain.MedicalRecord) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "appointment_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"diagnosis",
			"symptoms",
			"actions_to_avoid",
			"actions_to_take",
			"treatment_plan",
			"next_appointment_date",
			"next_appointment_note",
			"expert_notes",
			"updated_at",
		}),
	}).Create(record).Error
}

func (r *pgRepository) GetMedicalRecordByAppointmentID(appointmentID string) (*domain.MedicalRecord, error) {
	var record domain.MedicalRecord
	err := r.db.Where("appointment_id = ?", appointmentID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *pgRepository) GetMedicalRecordByID(recordID string) (*domain.MedicalRecord, error) {
	var record domain.MedicalRecord
	err := r.db.Where("record_id = ?", recordID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *pgRepository) ListMedicalRecordsByPatient(patientID string, limit, offset int) ([]domain.MedicalRecord, int64, error) {
	var records []domain.MedicalRecord
	var total int64

	query := r.db.Table(`"Booking_Medical_Records" mr`).
		Joins(`LEFT JOIN "Booking_Appointments" appt ON appt.appointment_id = mr.appointment_id`).
		Joins(`LEFT JOIN "Booking_Expert_Slots" slot ON slot.slot_id = appt.slot_id`)

	if patientID != "" {
		query = query.Where("mr.patient_id = ?", patientID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 10
	}

	err := query.Select("mr.*, slot.start_time as appointment_date, slot.price as appointment_price, appt.meeting_link, CASE appt.status WHEN 0 THEN 'PENDING_PAYMENT' WHEN 1 THEN 'CONFIRMED' WHEN 2 THEN 'CANCELLED' WHEN 3 THEN 'COMPLETED' ELSE 'UNKNOWN' END as appointment_status").
		Order("mr.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&records).Error

	return records, total, err
}

func (r *pgRepository) ListMedicalRecordsByExpert(expertID string, limit, offset int) ([]domain.MedicalRecord, int64, error) {
	var records []domain.MedicalRecord
	var total int64

	query := r.db.Table(`"Booking_Medical_Records" mr`).
		Joins(`LEFT JOIN "Booking_Appointments" appt ON appt.appointment_id = mr.appointment_id`).
		Joins(`LEFT JOIN "Booking_Expert_Slots" slot ON slot.slot_id = appt.slot_id`)

	if expertID != "" {
		query = query.Where("mr.expert_id = ?", expertID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 10
	}

	err := query.Select("mr.*, slot.start_time as appointment_date, slot.price as appointment_price, appt.meeting_link, CASE appt.status WHEN 0 THEN 'PENDING_PAYMENT' WHEN 1 THEN 'CONFIRMED' WHEN 2 THEN 'CANCELLED' WHEN 3 THEN 'COMPLETED' ELSE 'UNKNOWN' END as appointment_status").
		Order("mr.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&records).Error

	return records, total, err
}
