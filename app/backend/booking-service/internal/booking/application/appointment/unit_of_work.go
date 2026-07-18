package appointment

import (
	"context"

	"booking-service/internal/domain"
)

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	LoadAppointmentForUpdate(ctx context.Context, appointmentID string) (*domain.Appointment, error)
	LoadSlotForUpdate(ctx context.Context, slotID string) (*domain.ExpertSlot, error)
	CreateAppointment(ctx context.Context, appointment *domain.Appointment) error
	UpdateAppointment(ctx context.Context, appointment *domain.Appointment, updates map[string]interface{}) error
	UpdateSlot(ctx context.Context, slotID string, expectedStatus *domain.SlotStatus, updates map[string]interface{}) (int64, error)
}
