package appointment

import (
	"booking-service/internal/domain"

	"github.com/google/uuid"
)

func (u *appointmentUsecase) CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error) {
	appointment := &domain.Appointment{
		AppointmentID: uuid.New().String(),
		SlotID:        slotID,
		PatientID:     patientID,
		ExpertID:      expertID,
		Status:        domain.AppointmentStatusPendingPayment,
	}

	if err := u.repo.CreateAppointment(appointment); err != nil {
		return nil, err
	}
	return appointment, nil
}
