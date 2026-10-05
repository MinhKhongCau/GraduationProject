package payment

import (
	"context"
	"errors"
	paymentdomain "payment-service/internal/domain/payment"

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

type BookingDeliveryFailureCategory = paymentdomain.BookingDeliveryFailureCategory

const (
	BookingDeliveryNetwork           = paymentdomain.BookingDeliveryNetwork
	BookingDeliveryTimeout           = paymentdomain.BookingDeliveryTimeout
	BookingDeliveryRateLimited       = paymentdomain.BookingDeliveryRateLimited
	BookingDeliveryUpstream          = paymentdomain.BookingDeliveryUpstream
	BookingDeliveryMalformedResponse = paymentdomain.BookingDeliveryMalformedResponse
	BookingDeliveryNotFound          = paymentdomain.BookingDeliveryNotFound
	BookingDeliveryConflict          = paymentdomain.BookingDeliveryConflict
	BookingDeliveryAuthentication    = paymentdomain.BookingDeliveryAuthentication
	BookingDeliveryBadRequest        = paymentdomain.BookingDeliveryBadRequest
	BookingDeliveryBusinessRejection = paymentdomain.BookingDeliveryBusinessRejection
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
