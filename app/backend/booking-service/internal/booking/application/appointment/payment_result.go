package appointment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"booking-service/internal/booking/domain"
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
	if u.uow == nil {
		return u.repo.HandlePaymentResult(command)
	}

	return u.uow.WithinTx(context.Background(), func(tx Tx) error {
		appt, err := tx.LoadAppointmentForUpdate(context.Background(), command.AppointmentID)
		if err != nil {
			return err
		}

		slot, err := tx.LoadSlotForUpdate(context.Background(), appt.SlotID)
		if err != nil {
			if errors.Is(err, errTxRecordNotFound) {
				return fmt.Errorf("%w: slot not found", ErrPaymentResultConflict)
			}
			return err
		}

		transition, err := domain.PlanPaymentResultTransition(*appt, *slot, domain.PaymentResultStatus(command.Status), time.Now().UnixMilli())
		if err != nil {
			return mapPaymentResultDomainError(err)
		}
		if command.Status == PaymentResultFailed && !transition.Noop {
			covered, err := tx.IsSlotCoveredByTimeOff(context.Background(), *slot)
			if err != nil {
				return err
			}
			if covered {
				transition.SlotUpdates["status"] = domain.SlotStatusUnavailable
			}
		}
		return persistPaymentResultTransition(context.Background(), tx, appt, transition)
	})
}

func persistPaymentResultTransition(ctx context.Context, tx Tx, appt *domain.Appointment, transition *domain.PaymentResultTransition) error {
	if transition.Noop {
		return nil
	}
	if err := tx.UpdateAppointment(ctx, appt, transition.AppointmentUpdates); err != nil {
		return err
	}

	rowsAffected, err := tx.UpdateSlot(ctx, appt.SlotID, &transition.ExpectedSlotStatus, transition.SlotUpdates)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: slot is no longer %s", ErrPaymentResultConflict, transition.ExpectedSlotStatus.String())
	}
	return nil
}

func mapPaymentResultDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidPaymentResultStatus):
		return fmt.Errorf("%w: %v", ErrInvalidPaymentResultStatus, err)
	case errors.Is(err, domain.ErrPaymentResultConflict):
		return fmt.Errorf("%w: %v", ErrPaymentResultConflict, err)
	default:
		return err
	}
}
