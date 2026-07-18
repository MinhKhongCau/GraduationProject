package appointment

import "booking-service/internal/domain"

func (u *appointmentUsecase) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return u.repo.GetAppointmentByID(appointmentID)
}

func (u *appointmentUsecase) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByPatient(patientID)
}

func (u *appointmentUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByExpert(expertID, fromDate, toDate, status)
}
