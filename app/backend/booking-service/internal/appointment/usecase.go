package appointment

import (
	"booking-service/internal/domain"
	"github.com/google/uuid"
)

type Usecase interface {
	CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error)
	CancelAppointment(appointmentID, userID, userRole, reason string) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentFailure(appointmentID string) error
	GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error)
}

type appointmentUsecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &appointmentUsecase{repo: repo}
}

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

func (u *appointmentUsecase) CancelAppointment(appointmentID, userID, userRole, reason string) error {
	if userRole == "PATIENT" {
		// TODO: Mốc thời gian huỷ tối thiểu (policy)
		// Check nếu cuộc hẹn bắt đầu trong vòng 24h thì chặn không cho huỷ
		return u.repo.CancelAppointmentByPatient(appointmentID, userID, reason)
	} else if userRole == "EXPERT" {
		return u.repo.CancelAppointmentByExpert(appointmentID, reason)
	}
	return ErrUnauthorized
}

func (u *appointmentUsecase) ConfirmPayment(appointmentID string) error {
	return u.repo.ConfirmPayment(appointmentID)
}

func (u *appointmentUsecase) HandlePaymentFailure(appointmentID string) error {
	// Let worker clean up the slot naturally or unlock it immediately if needed
	// In the webhook handler, it just ignores or logs, but we could explicitly cancel here.
	return nil
}

func (u *appointmentUsecase) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByPatient(patientID)
}

func (u *appointmentUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByExpert(expertID, fromDate, toDate, status)
}
