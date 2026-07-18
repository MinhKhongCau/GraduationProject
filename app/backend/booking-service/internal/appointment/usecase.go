package appointment

import "booking-service/internal/domain"

type Usecase interface {
	CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error)
	GetAppointmentByID(appointmentID string) (*domain.Appointment, error)
	GetPaymentEligibility(command GetPaymentEligibilityCommand) (PaymentEligibility, error)
	CancelAppointment(appointmentID, userID, userRole, reason string) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentFailure(appointmentID string) error
	HandlePaymentResult(command HandlePaymentResultCommand) error
	GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error)
}

type appointmentUsecase struct {
	repo Repository
	uow  UnitOfWork
}

func NewUsecase(repo Repository) Usecase {
	usecase := &appointmentUsecase{repo: repo}
	if uow, ok := repo.(UnitOfWork); ok {
		usecase.uow = uow
	}
	return usecase
}
