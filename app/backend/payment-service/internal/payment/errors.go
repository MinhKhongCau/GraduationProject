package payment

import apppayment "payment-service/internal/payment/application"

var (
	ErrInvalidCreateOrderRequest  = apppayment.ErrInvalidCreateOrderRequest
	ErrUnsupportedGateway         = apppayment.ErrUnsupportedGateway
	ErrBookingAppointmentNotFound = apppayment.ErrBookingAppointmentNotFound
	ErrAppointmentOwnership       = apppayment.ErrAppointmentOwnership
	ErrAppointmentInvalidState    = apppayment.ErrAppointmentInvalidState
	ErrInvalidBookingData         = apppayment.ErrInvalidBookingData
)
