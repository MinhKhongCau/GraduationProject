package repository

import (
	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"context"
)

type meetingRow struct {
	AppointmentID      string
	PatientID          string
	ExpertID           string
	Status             appointmentdomain.AppointmentStatus
	StartTime          int64
	EndTime            int64
	SessionCompletedAt *int64
}

// GetMeetingByToken tra lịch hẹn (kèm giờ của slot) theo mã phòng họp trong link.
func (r *pgRepository) GetMeetingByToken(ctx context.Context, token string) (*appappointment.MeetingInfo, error) {
	var rows []meetingRow
	err := r.db.WithContext(ctx).Table(`"Booking_Appointments" appointment`).
		Joins(`JOIN "Booking_Expert_Slots" slot ON slot.slot_id = appointment.slot_id`).
		Where("appointment.meeting_token = ?", token).
		Select("appointment.appointment_id, appointment.patient_id, appointment.expert_id, appointment.status, slot.start_time, slot.end_time, appointment.session_completed_at").
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, appappointment.ErrMeetingNotFound
	}
	row := rows[0]
	return &appappointment.MeetingInfo{
		AppointmentID:      row.AppointmentID,
		PatientID:          row.PatientID,
		ExpertID:           row.ExpertID,
		Status:             row.Status.String(),
		StartTime:          row.StartTime,
		EndTime:            row.EndTime,
		SessionCompletedAt: row.SessionCompletedAt,
	}, nil
}
