package appointment

import (
	"booking-service/internal/domain"

	"gorm.io/gorm"
)

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
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}
