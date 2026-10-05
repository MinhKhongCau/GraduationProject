package appointment

import (
	bookingquery "booking-service/internal/application/query"
	appointmentdomain "booking-service/internal/domain/appointment"
)

type Usecase interface {
	CreateAppointment(patientID, expertID, slotID string) (*appointmentdomain.Appointment, error)
	GetAppointmentByID(appointmentID string) (*appointmentdomain.Appointment, error)
	GetPaymentEligibility(command GetPaymentEligibilityCommand) (PaymentEligibility, error)
	CancelAppointment(appointmentID, userID, userRole, reason string) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentFailure(appointmentID string) error
	HandlePaymentResult(command HandlePaymentResultCommand) error
	GetAppointmentsByPatient(patientID string) ([]appointmentdomain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *appointmentdomain.AppointmentStatus) ([]appointmentdomain.Appointment, error)

	SaveMedicalRecord(actorID, actorRole, appointmentID string, cmd SaveMedicalRecordCommand) (*appointmentdomain.MedicalRecord, error)
	GetMedicalRecordByAppointmentID(actorID, actorRole, appointmentID string) (*appointmentdomain.MedicalRecord, error)
	GetMedicalRecordByID(actorID, actorRole, recordID string) (*appointmentdomain.MedicalRecord, error)
	ListMedicalRecords(actorID, actorRole string, page bookingquery.PageRequest) (bookingquery.Page[appointmentdomain.MedicalRecord], error)
}

type ReadUsecase interface {
	ListAppointments(query AppointmentListQuery) (bookingquery.Page[appointmentdomain.Appointment], error)
	GetAppointmentDetail(actorID, actorRole, appointmentID string) (*appointmentdomain.Appointment, error)
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
