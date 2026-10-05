package appointment

import (
	"booking-service/internal/domain/shared"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"fmt"
)

var (
	ErrPaymentEligibilityForbidden = errors.New("appointment does not belong to payer")
	ErrPaymentEligibilityConflict  = errors.New("appointment is not eligible for payment")
)

type PaymentEligibility struct {
	AppointmentID string
	ExpertID      string
	AmountVND     shared.MoneyVND
	ExpiresAt     int64
}

func BuildPaymentEligibility(appointment Appointment, slot slotdomain.ExpertSlot, payerID string, nowMs int64) (PaymentEligibility, error) {
	if !shared.SameID(appointment.PatientID, payerID) {
		return PaymentEligibility{}, ErrPaymentEligibilityForbidden
	}
	if appointment.Status != AppointmentStatusPendingPayment {
		return PaymentEligibility{}, fmt.Errorf("%w: appointment status is %d", ErrPaymentEligibilityConflict, appointment.Status)
	}
	if slot.SlotID == "" {
		return PaymentEligibility{}, fmt.Errorf("%w: slot not found", ErrPaymentEligibilityConflict)
	}
	if !shared.SameID(slot.ExpertID, appointment.ExpertID) {
		return PaymentEligibility{}, fmt.Errorf("%w: slot expert does not match appointment expert", ErrPaymentEligibilityConflict)
	}
	if slot.Status != slotdomain.SlotStatusLocked {
		return PaymentEligibility{}, fmt.Errorf("%w: slot status is %d", ErrPaymentEligibilityConflict, slot.Status)
	}
	if slot.LockedBy == nil || !shared.SameID(*slot.LockedBy, payerID) {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock owner mismatch", ErrPaymentEligibilityConflict)
	}
	if slot.LockedExpiresAt == nil {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock expiry is missing", ErrPaymentEligibilityConflict)
	}
	if *slot.LockedExpiresAt <= nowMs {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock expired", ErrPaymentEligibilityConflict)
	}

	amountVND, err := shared.NewMoneyVNDFromPrice(slot.Price)
	if err != nil {
		return PaymentEligibility{}, err
	}

	return PaymentEligibility{
		AppointmentID: appointment.AppointmentID,
		ExpertID:      appointment.ExpertID,
		AmountVND:     amountVND,
		ExpiresAt:     *slot.LockedExpiresAt,
	}, nil
}
