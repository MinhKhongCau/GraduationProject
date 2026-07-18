package application

import "errors"

var (
	ErrInvalidCreateOrderRequest  = errors.New("invalid payment order request")
	ErrUnsupportedGateway         = errors.New("unsupported payment gateway")
	ErrBookingAppointmentNotFound = errors.New("booking appointment not found")
	ErrAppointmentOwnership       = errors.New("appointment does not belong to payer")
	ErrAppointmentInvalidState    = errors.New("appointment is not pending payment")
	ErrInvalidBookingData         = errors.New("invalid booking appointment data")
	ErrTxRecordNotFound           = errors.New("record not found")
)
