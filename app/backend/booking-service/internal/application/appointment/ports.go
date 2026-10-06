package appointment

import (
	"context"

	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
)

type Repository interface {
	GetAppointmentByID(appointmentID string) (*appointmentdomain.Appointment, error)
	GetPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (*PaymentEligibilitySnapshot, error)
	GetAppointmentBySlotID(slotID string) (*appointmentdomain.Appointment, error)
	CancelAppointmentByExpert(appointmentID string, reason string) error
	CancelAppointmentByPatient(appointmentID string, patientID string, reason string) error
	GetAppointmentsByPatient(patientID string) ([]appointmentdomain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *appointmentdomain.AppointmentStatus) ([]appointmentdomain.Appointment, error)
	LockSlot(slotID string, patientID string) error
	CreateAppointment(appointment *appointmentdomain.Appointment) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentResult(command HandlePaymentResultCommand) error
	CancelExpiredLocks() (int64, error)
	UpdateAppointmentStatus(appointmentID string, status appointmentdomain.AppointmentStatus) error

	SaveMedicalRecord(record *appointmentdomain.MedicalRecord) error
	GetMedicalRecordByAppointmentID(appointmentID string) (*appointmentdomain.MedicalRecord, error)
	GetMedicalRecordByID(recordID string) (*appointmentdomain.MedicalRecord, error)
	ListMedicalRecordsByPatient(patientID string, limit, offset int) ([]appointmentdomain.MedicalRecord, int64, error)
	ListMedicalRecordsByExpert(expertID string, limit, offset int) ([]appointmentdomain.MedicalRecord, int64, error)
}

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	LoadAppointmentForUpdate(ctx context.Context, appointmentID string) (*appointmentdomain.Appointment, error)
	LoadSlotForUpdate(ctx context.Context, slotID string) (*slotdomain.ExpertSlot, error)
	IsSlotCoveredByTimeOff(ctx context.Context, slot slotdomain.ExpertSlot) (bool, error)
	// HasActiveAppointmentForSlot: slot đã có cuộc hẹn PENDING_PAYMENT/CONFIRMED.
	HasActiveAppointmentForSlot(ctx context.Context, slotID string) (bool, error)
	CreateAppointment(ctx context.Context, appointment *appointmentdomain.Appointment) error
	UpdateAppointment(ctx context.Context, appointment *appointmentdomain.Appointment, updates map[string]interface{}) error
	UpdateSlot(ctx context.Context, slotID string, expectedStatus *slotdomain.SlotStatus, updates map[string]interface{}) (int64, error)
}
