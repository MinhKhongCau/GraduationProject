package appointment

import (
	"booking-service/internal/domain"
	"context"
	"errors"
	"time"

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

	if err := u.createAppointmentWithUOW(context.Background(), appointment); err != nil {
		return nil, err
	}
	return appointment, nil
}

func (u *appointmentUsecase) createAppointmentWithUOW(ctx context.Context, appointment *domain.Appointment) error {
	if u.uow == nil {
		return u.repo.CreateAppointment(appointment)
	}

	return u.uow.WithinTx(ctx, func(tx Tx) error {
		slot, err := tx.LoadSlotForUpdate(ctx, appointment.SlotID)
		if err != nil {
			return errors.New("khÃ´ng tÃ¬m tháº¥y slot: " + err.Error())
		}

		nowMs := time.Now().UnixMilli()
		if err := domain.ValidateAppointmentCreationSlot(*slot, appointment, nowMs); err != nil {
			return mapAppointmentCreationSlotError(err)
		}

		appointment.CreatedAt = nowMs
		appointment.UpdatedAt = nowMs
		if err := tx.CreateAppointment(ctx, appointment); err != nil {
			return errors.New("lá»—i khi táº¡o cuá»™c háº¹n: " + err.Error())
		}

		return nil
	})
}
