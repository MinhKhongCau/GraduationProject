package appointment

import "booking-service/internal/domain"

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

type PaymentEligibilitySnapshot struct {
	Appointment domain.Appointment
	Slot        domain.ExpertSlot
}
