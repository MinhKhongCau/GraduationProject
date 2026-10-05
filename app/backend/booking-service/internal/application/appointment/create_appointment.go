package appointment

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (u *appointmentUsecase) CreateAppointment(patientID, expertID, slotID string) (*appointmentdomain.Appointment, error) {
	appointment := &appointmentdomain.Appointment{
		AppointmentID: uuid.New().String(),
		SlotID:        slotID,
		PatientID:     patientID,
		ExpertID:      expertID,
		Status:        appointmentdomain.AppointmentStatusPendingPayment,
	}

	if err := u.createAppointmentWithUOW(context.Background(), appointment); err != nil {
		return nil, err
	}
	return appointment, nil
}

func (u *appointmentUsecase) createAppointmentWithUOW(ctx context.Context, appointment *appointmentdomain.Appointment) error {
	return CreateAppointmentWithUnitOfWork(ctx, u.repo, u.uow, appointment)
}

// CreateAppointmentWithUnitOfWork exists temporarily so the old postgres adapter
// can preserve its compatibility CreateAppointment method until that adapter moves.
func CreateAppointmentWithUnitOfWork(ctx context.Context, repo Repository, uow UnitOfWork, appointment *appointmentdomain.Appointment) error {
	if uow == nil {
		return repo.CreateAppointment(appointment)
	}

	return uow.WithinTx(ctx, func(tx Tx) error {
		slot, err := tx.LoadSlotForUpdate(ctx, appointment.SlotID)
		if err != nil {
			return errors.New("khÃ´ng tÃ¬m tháº¥y slot: " + err.Error())
		}

		nowMs := time.Now().UnixMilli()
		if err := appointmentdomain.ValidateAppointmentCreationSlot(*slot, appointment, nowMs); err != nil {
			return mapAppointmentCreationSlotError(err)
		}
		covered, err := tx.IsSlotCoveredByTimeOff(ctx, *slot)
		if err != nil {
			return err
		}
		if covered {
			return errors.New("slot is covered by expert time-off")
		}

		appointment.CreatedAt = nowMs
		appointment.UpdatedAt = nowMs
		if err := tx.CreateAppointment(ctx, appointment); err != nil {
			return errors.New("lá»—i khi táº¡o cuá»™c háº¹n: " + err.Error())
		}

		return nil
	})
}

func mapAppointmentCreationSlotError(err error) error {
	switch {
	case errors.Is(err, appointmentdomain.ErrAppointmentCreationSlotNotLockedByPatient):
		return errors.New("slot khÃ´ng Ä‘Æ°á»£c giá»¯ bá»Ÿi báº¡n, vui lÃ²ng thá»±c hiá»‡n láº¡i tá»« Ä‘áº§u")
	case errors.Is(err, appointmentdomain.ErrAppointmentCreationSlotExpertMismatch):
		return errors.New("slot expert does not match appointment expert")
	case errors.Is(err, appointmentdomain.ErrAppointmentCreationSlotLockExpired):
		return errors.New("phiÃªn giá»¯ chá»— Ä‘Ã£ háº¿t háº¡n 15 phÃºt, vui lÃ²ng chá»n láº¡i")
	default:
		return err
	}
}
