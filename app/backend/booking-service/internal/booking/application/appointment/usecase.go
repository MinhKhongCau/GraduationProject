package appointment

import (
	bookingquery "booking-service/internal/booking/application/query"
	"booking-service/internal/booking/domain"
)

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

	SaveMedicalRecord(actorID, actorRole, appointmentID string, cmd SaveMedicalRecordCommand) (*domain.MedicalRecord, error)
	GetMedicalRecordByAppointmentID(actorID, actorRole, appointmentID string) (*domain.MedicalRecord, error)
	GetMedicalRecordByID(actorID, actorRole, recordID string) (*domain.MedicalRecord, error)
	ListMedicalRecords(actorID, actorRole string, page bookingquery.PageRequest) (bookingquery.Page[domain.MedicalRecord], error)
}

type ReadUsecase interface {
	ListAppointments(query AppointmentListQuery) (bookingquery.Page[domain.Appointment], error)
	GetAppointmentDetail(actorID, actorRole, appointmentID string) (*domain.Appointment, error)
}

type Repository interface {
	GetAppointmentByID(appointmentID string) (*domain.Appointment, error)
	GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error)
	GetAppointmentBySlotID(slotID string) (*domain.Appointment, error)
	CancelAppointmentByExpert(appointmentID string, reason string) error
	CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error
	GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error)
	LockSlot(slotID string, patientID string) error
	CreateAppointment(appointment *domain.Appointment) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentResult(command HandlePaymentResultCommand) error
	CancelExpiredLocks() (int64, error)
	UpdateAppointmentStatus(appointmentID string, status domain.AppointmentStatus) error

	SaveMedicalRecord(record *domain.MedicalRecord) error
	GetMedicalRecordByAppointmentID(appointmentID string) (*domain.MedicalRecord, error)
	GetMedicalRecordByID(recordID string) (*domain.MedicalRecord, error)
	ListMedicalRecordsByPatient(patientID string, limit, offset int) ([]domain.MedicalRecord, int64, error)
	ListMedicalRecordsByExpert(expertID string, limit, offset int) ([]domain.MedicalRecord, int64, error)
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
