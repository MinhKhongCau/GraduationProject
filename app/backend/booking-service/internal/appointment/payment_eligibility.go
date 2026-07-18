package appointment

import (
	"booking-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type GetPaymentEligibilityCommand struct {
	AppointmentID string
	PayerID       string
}

type PaymentEligibility struct {
	AppointmentID string
	ExpertID      string
	AmountVND     int64
	ExpiresAt     int64
}

func (u *appointmentUsecase) GetPaymentEligibility(command GetPaymentEligibilityCommand) (PaymentEligibility, error) {
	command.AppointmentID = strings.TrimSpace(command.AppointmentID)
	command.PayerID = strings.TrimSpace(command.PayerID)
	if command.AppointmentID == "" {
		return PaymentEligibility{}, ErrNotFound
	}
	if command.PayerID == "" {
		return PaymentEligibility{}, ErrPaymentEligibilityForbidden
	}

	eligibility, err := u.getPaymentEligibilitySnapshot(command)
	if err != nil {
		return PaymentEligibility{}, err
	}

	return eligibility, nil
}

func mapPaymentEligibilityDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrPaymentEligibilityForbidden):
		return fmt.Errorf("%w: %v", ErrPaymentEligibilityForbidden, err)
	case errors.Is(err, domain.ErrPaymentEligibilityConflict):
		return fmt.Errorf("%w: %v", ErrPaymentEligibilityConflict, err)
	case errors.Is(err, domain.ErrInvalidMoneyVND):
		return fmt.Errorf("%w: %v", ErrInvalidBookingPrice, err)
	default:
		return err
	}
}

func (u *appointmentUsecase) getPaymentEligibilitySnapshot(command GetPaymentEligibilityCommand) (PaymentEligibility, error) {
	if u.uow == nil {
		snapshot, err := u.repo.GetPaymentEligibilitySnapshot(command)
		if err != nil {
			return PaymentEligibility{}, err
		}
		return buildApplicationPaymentEligibility(snapshot.Appointment, snapshot.Slot, command.PayerID)
	}

	var result PaymentEligibility
	err := u.uow.WithinTx(context.Background(), func(tx Tx) error {
		appt, err := tx.LoadAppointmentForUpdate(context.Background(), command.AppointmentID)
		if err != nil {
			return err
		}

		slot, err := tx.LoadSlotForUpdate(context.Background(), appt.SlotID)
		if err != nil {
			if errors.Is(err, errTxRecordNotFound) {
				return fmt.Errorf("%w: slot not found", ErrPaymentEligibilityConflict)
			}
			return err
		}

		eligibility, err := buildApplicationPaymentEligibility(*appt, *slot, command.PayerID)
		if err != nil {
			return err
		}
		result = eligibility
		return nil
	})
	if err != nil {
		return PaymentEligibility{}, err
	}
	return result, nil
}

func buildApplicationPaymentEligibility(appt domain.Appointment, slot domain.ExpertSlot, payerID string) (PaymentEligibility, error) {
	eligibility, err := domain.BuildPaymentEligibility(appt, slot, payerID, time.Now().UnixMilli())
	if err != nil {
		return PaymentEligibility{}, mapPaymentEligibilityDomainError(err)
	}

	return PaymentEligibility{
		AppointmentID: eligibility.AppointmentID,
		ExpertID:      eligibility.ExpertID,
		AmountVND:     int64(eligibility.AmountVND),
		ExpiresAt:     eligibility.ExpiresAt,
	}, nil
}
