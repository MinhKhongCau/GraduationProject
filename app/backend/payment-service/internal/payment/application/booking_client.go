package application

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

type BookingDeliveryFailureCategory string

const (
	BookingDeliveryNetwork           BookingDeliveryFailureCategory = "network"
	BookingDeliveryTimeout           BookingDeliveryFailureCategory = "timeout"
	BookingDeliveryRateLimited       BookingDeliveryFailureCategory = "rate_limited"
	BookingDeliveryUpstream          BookingDeliveryFailureCategory = "upstream"
	BookingDeliveryMalformedResponse BookingDeliveryFailureCategory = "malformed_response"
	BookingDeliveryNotFound          BookingDeliveryFailureCategory = "not_found"
	BookingDeliveryConflict          BookingDeliveryFailureCategory = "conflict"
	BookingDeliveryAuthentication    BookingDeliveryFailureCategory = "authentication"
	BookingDeliveryBadRequest        BookingDeliveryFailureCategory = "bad_request"
	BookingDeliveryBusinessRejection BookingDeliveryFailureCategory = "business_rejection"
)

type BookingDeliveryError struct {
	Category  BookingDeliveryFailureCategory
	Retryable bool
	Message   string
}

func (e *BookingDeliveryError) Error() string {
	return e.Message
}

type BookingServiceClient interface {
	GetPaymentEligibility(ctx context.Context, appointmentID string, payerID uuid.UUID) (*PaymentEligibility, error)
	ConfirmAppointment(ctx context.Context, appointmentID string) error
	FailAppointment(ctx context.Context, appointmentID string) error
}
