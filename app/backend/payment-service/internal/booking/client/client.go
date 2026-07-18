// Package client defines the port Payment Service uses to communicate with
// Booking Service. REST is the current adapter; a future gRPC adapter should
// return the same application contract.
package client

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrAppointmentNotFound         = errors.New("booking appointment not found")
	ErrPaymentEligibilityForbidden = errors.New("booking payment eligibility forbidden")
	ErrPaymentEligibilityConflict  = errors.New("booking payment eligibility conflict")
	ErrInvalidBookingPrice         = errors.New("invalid booking price")
)

type PaymentEligibility struct {
	AppointmentID string
	ExpertID      string
	AmountVND     int64
	ExpiresAt     int64
}

type BookingServiceClient interface {
	GetPaymentEligibility(ctx context.Context, appointmentID string, payerID uuid.UUID) (*PaymentEligibility, error)
	ConfirmAppointment(ctx context.Context, appointmentID string) error
	FailAppointment(ctx context.Context, appointmentID string) error
}
