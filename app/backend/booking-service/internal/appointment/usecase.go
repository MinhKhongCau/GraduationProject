package appointment

import (
	"booking-service/internal/domain"
	"strings"

	"github.com/google/uuid"
)

type PaymentResultStatus string

const (
	PaymentResultSuccess PaymentResultStatus = "SUCCESS"
	PaymentResultFailed  PaymentResultStatus = "FAILED"
)

type HandlePaymentResultCommand struct {
	AppointmentID string
	Status        PaymentResultStatus
}

type Usecase interface {
	CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error)
	GetAppointmentByID(appointmentID string) (*domain.Appointment, error)
	CancelAppointment(appointmentID, userID, userRole, reason string) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentFailure(appointmentID string) error
	HandlePaymentResult(command HandlePaymentResultCommand) error
	GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error)
}

type appointmentUsecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &appointmentUsecase{repo: repo}
}

func ParsePaymentResultStatus(status string) (PaymentResultStatus, error) {
	switch PaymentResultStatus(strings.TrimSpace(status)) {
	case PaymentResultSuccess:
		return PaymentResultSuccess, nil
	case PaymentResultFailed:
		return PaymentResultFailed, nil
	default:
		return "", ErrInvalidPaymentResultStatus
	}
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

func (u *appointmentUsecase) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return u.repo.GetAppointmentByID(appointmentID)
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
	return u.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultSuccess,
	})
}

func (u *appointmentUsecase) HandlePaymentFailure(appointmentID string) error {
	return u.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultFailed,
	})
}

func (u *appointmentUsecase) HandlePaymentResult(command HandlePaymentResultCommand) error {
	if command.AppointmentID == "" {
		return ErrNotFound
	}
	if command.Status != PaymentResultSuccess && command.Status != PaymentResultFailed {
		return ErrInvalidPaymentResultStatus
	}
	return u.repo.HandlePaymentResult(command)
}

func (u *appointmentUsecase) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByPatient(patientID)
}

func (u *appointmentUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByExpert(expertID, fromDate, toDate, status)
}
