package payment

import "errors"

var (
	ErrInvalidCreateOrderRequest     = errors.New("invalid payment order request")
	ErrUnsupportedGateway            = errors.New("unsupported payment gateway")
	ErrBookingAppointmentNotFound    = errors.New("booking appointment not found")
	ErrAppointmentOwnership          = errors.New("appointment does not belong to payer")
	ErrAppointmentInvalidState       = errors.New("appointment is not pending payment")
	ErrInvalidBookingData            = errors.New("invalid booking appointment data")
	ErrTxRecordNotFound              = errors.New("record not found")
	ErrAppointmentAlreadyPaid        = errors.New("appointment is already paid")
	ErrPaymentWindowTooShort         = errors.New("payment window is too short")
	ErrExistingOrderConflict         = errors.New("existing payment order conflicts with booking eligibility")
	ErrActivePendingOrderExists      = errors.New("active pending payment order already exists")
	ErrInvalidCompensationFilter     = errors.New("invalid compensation case filter")
	ErrCompensationCaseNotFound      = errors.New("payment compensation case not found")
	ErrCompensationReaderUnavailable = errors.New("compensation case reader is unavailable")
	ErrInvalidPaymentOrderFilter     = errors.New("invalid payment order filter")
	ErrPaymentOrderNotFound          = errors.New("payment order not found")
	ErrPaymentOrderForbidden         = errors.New("payment order access denied")
	ErrPaymentOrderReaderUnavailable = errors.New("payment order reader unavailable")
)
