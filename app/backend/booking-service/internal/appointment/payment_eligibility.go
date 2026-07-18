package appointment

import (
	"booking-service/internal/domain"
	"fmt"
	"math"
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

	return buildPaymentEligibility(command, snapshot.Appointment, snapshot.Slot, time.Now().UnixMilli())
}

func buildPaymentEligibility(command GetPaymentEligibilityCommand, appt domain.Appointment, slot domain.ExpertSlot, nowMs int64) (PaymentEligibility, error) {
	if !sameID(appt.PatientID, command.PayerID) {
		return PaymentEligibility{}, ErrPaymentEligibilityForbidden
	}
	if appt.Status != domain.AppointmentStatusPendingPayment {
		return PaymentEligibility{}, fmt.Errorf("%w: appointment status is %d", ErrPaymentEligibilityConflict, appt.Status)
	}
	if slot.SlotID == "" {
		return PaymentEligibility{}, fmt.Errorf("%w: slot not found", ErrPaymentEligibilityConflict)
	}
	if !sameID(slot.ExpertID, appt.ExpertID) {
		return PaymentEligibility{}, fmt.Errorf("%w: slot expert does not match appointment expert", ErrPaymentEligibilityConflict)
	}
	if slot.Status != domain.SlotStatusLocked {
		return PaymentEligibility{}, fmt.Errorf("%w: slot status is %d", ErrPaymentEligibilityConflict, slot.Status)
	}
	if slot.LockedBy == nil || !sameID(*slot.LockedBy, command.PayerID) {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock owner mismatch", ErrPaymentEligibilityConflict)
	}
	if slot.LockedExpiresAt == nil {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock expiry is missing", ErrPaymentEligibilityConflict)
	}
	if *slot.LockedExpiresAt <= nowMs {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock expired", ErrPaymentEligibilityConflict)
	}

	amountVND, err := strictPriceToVND(slot.Price)
	if err != nil {
		return PaymentEligibility{}, err
	}

	return PaymentEligibility{
		AppointmentID: appt.AppointmentID,
		ExpertID:      appt.ExpertID,
		AmountVND:     amountVND,
		ExpiresAt:     *slot.LockedExpiresAt,
	}, nil
}

func strictPriceToVND(price float64) (int64, error) {
	if math.IsNaN(price) {
		return 0, fmt.Errorf("%w: price is NaN", ErrInvalidBookingPrice)
	}
	if math.IsInf(price, 0) {
		return 0, fmt.Errorf("%w: price is infinite", ErrInvalidBookingPrice)
	}
	if price <= 0 {
		return 0, fmt.Errorf("%w: price must be greater than zero", ErrInvalidBookingPrice)
	}
	if math.Trunc(price) != price {
		return 0, fmt.Errorf("%w: fractional VND is not supported", ErrInvalidBookingPrice)
	}
	const maxVNPaySafeAmountVND = float64(92233720368547758)
	if price >= maxVNPaySafeAmountVND {
		return 0, fmt.Errorf("%w: price is too large", ErrInvalidBookingPrice)
	}
	return int64(price), nil
}

func sameID(left, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}
