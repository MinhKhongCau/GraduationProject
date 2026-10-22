package repository

import (
	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"errors"

	"gorm.io/gorm"
)

// Láº¥y Appointment theo SlotID (Æ°u tiÃªn láº¥y cuá»™c háº¹n chÆ°a bá»‹ huá»·, náº¿u khÃ´ng thÃ¬ láº¥y cuá»™c háº¹n má»›i nháº¥t)
func (r *pgRepository) GetAppointmentBySlotID(slotID string) (*appointmentdomain.Appointment, error) {
	var appt appointmentdomain.Appointment
	err := r.db.Where("slot_id = ? AND status != ?", slotID, appointmentdomain.AppointmentStatusCancelled).Order("created_at DESC").First(&appt).Error
	if err != nil {
		err = r.db.Where("slot_id = ?", slotID).Order("created_at DESC").First(&appt).Error
		if err != nil {
			return nil, err
		}
	}
	return &appt, nil
}

// Láº¥y Appointment theo AppointmentID
func (r *pgRepository) GetAppointmentByID(appointmentID string) (*appointmentdomain.Appointment, error) {
	var appt appointmentdomain.Appointment
	err := r.db.Table("Booking_Appointments").
		Select("\"Booking_Appointments\".*, \"Booking_Expert_Slots\".price, \"Booking_Expert_Slots\".start_time, \"Booking_Expert_Slots\".end_time, "+hasMedicalRecordColumn(`"Booking_Appointments"`)).
		Joins("JOIN \"Booking_Expert_Slots\" ON \"Booking_Appointments\".slot_id = \"Booking_Expert_Slots\".slot_id").
		Where("\"Booking_Appointments\".appointment_id = ?", appointmentID).
		First(&appt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appappointment.ErrNotFound
		}
		return nil, err
	}
	return &appt, nil
}

func (r *pgRepository) ListAppointments(filter appappointment.AppointmentListQuery) ([]appointmentdomain.Appointment, int64, error) {
	query := r.db.Table(`"Booking_Appointments" appointment`).Joins(`JOIN "Booking_Expert_Slots" slot ON slot.slot_id = appointment.slot_id`)
	if filter.PatientID != "" {
		query = query.Where("appointment.patient_id = ?", filter.PatientID)
	}
	if filter.ExpertID != "" {
		query = query.Where("appointment.expert_id = ?", filter.ExpertID)
	}
	if filter.ScopeExperts {
		query = query.Where("appointment.expert_id IN ?", filter.ExpertIDs)
	}
	if filter.FromMs > 0 {
		query = query.Where("slot.start_time >= ?", filter.FromMs)
	}
	if filter.ToMs > 0 {
		query = query.Where("slot.start_time < ?", filter.ToMs)
	}
	if filter.Status != nil {
		query = query.Where("appointment.status = ?", *filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var appointments []appointmentdomain.Appointment
	err := query.Select("appointment.*, slot.price, slot.start_time, slot.end_time, " + hasMedicalRecordColumn("appointment")).Order(appointmentOrder(filter.Sort)).Limit(filter.Page.Size).Offset(filter.Page.Offset()).Scan(&appointments).Error
	for i := range appointments {
		appointments[i].StatusLabel = appointments[i].Status.String()
	}
	return appointments, total, err
}

// hasMedicalRecordColumn: cột has_medical_record = lịch hẹn đã có bệnh án (không tải nội dung bệnh án).
func hasMedicalRecordColumn(appointmentTable string) string {
	return `EXISTS (SELECT 1 FROM "Booking_Medical_Records" record WHERE record.appointment_id = ` + appointmentTable + `.appointment_id) AS has_medical_record`
}

// appointmentOrder dựng ORDER BY từ lựa chọn đã kiểm tra (không nhận chuỗi tự do từ request);
// appointment_id làm khoá phụ để phân trang ổn định.
func appointmentOrder(sort appappointment.AppointmentSort) string {
	column := "slot.start_time"
	if sort.By == appappointment.AppointmentSortByCreatedAt {
		column = "appointment.created_at"
	}
	direction := "DESC"
	if sort.Ascending {
		direction = "ASC"
	}
	return column + " " + direction + ", appointment.appointment_id " + direction
}

// GetAppointmentsByIDs đọc nhiều lịch hẹn kèm giá/giờ khám của slot; id không tồn tại bị bỏ qua.
func (r *pgRepository) GetAppointmentsByIDs(appointmentIDs []string) ([]appointmentdomain.Appointment, error) {
	var appointments []appointmentdomain.Appointment
	err := r.db.Table(`"Booking_Appointments" appointment`).
		Joins(`JOIN "Booking_Expert_Slots" slot ON slot.slot_id = appointment.slot_id`).
		Where("appointment.appointment_id IN ?", appointmentIDs).
		Select("appointment.*, slot.price, slot.start_time, slot.end_time, " + hasMedicalRecordColumn("appointment")).
		Scan(&appointments).Error
	for i := range appointments {
		appointments[i].StatusLabel = appointments[i].Status.String()
	}
	return appointments, err
}

// Láº¥y danh sÃ¡ch cuá»™c háº¹n cá»§a Patient
func (r *pgRepository) GetAppointmentsByPatient(patientID string) ([]appointmentdomain.Appointment, error) {
	var appts []appointmentdomain.Appointment
	err := r.db.Where("patient_id = ?", patientID).Order("created_at desc").Find(&appts).Error
	return appts, err
}

// Láº¥y danh sÃ¡ch cuá»™c háº¹n cá»§a Expert
func (r *pgRepository) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *appointmentdomain.AppointmentStatus) ([]appointmentdomain.Appointment, error) {
	var appts []appointmentdomain.Appointment
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
