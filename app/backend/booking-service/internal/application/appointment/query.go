package appointment

import (
	appointmentdomain "booking-service/internal/domain/appointment"
)

func (u *appointmentUsecase) GetAppointmentByID(appointmentID string) (*appointmentdomain.Appointment, error) {
	return u.repo.GetAppointmentByID(appointmentID)
}

func (u *appointmentUsecase) GetAppointmentsByPatient(patientID string) ([]appointmentdomain.Appointment, error) {
	return u.repo.GetAppointmentsByPatient(patientID)
}

func (u *appointmentUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *appointmentdomain.AppointmentStatus) ([]appointmentdomain.Appointment, error) {
	return u.repo.GetAppointmentsByExpert(expertID, fromDate, toDate, status)
}
