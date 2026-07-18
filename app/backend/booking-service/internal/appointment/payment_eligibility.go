package appointment

import (
	"booking-service/internal/domain"
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

	snapshot, err := u.repo.GetPaymentEligibilitySnapshot(command)
	if err != nil {
		return PaymentEligibility{}, err
	}

	eligibility, err := domain.BuildPaymentEligibility(snapshot.Appointment, snapshot.Slot, command.PayerID, time.Now().UnixMilli())
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
