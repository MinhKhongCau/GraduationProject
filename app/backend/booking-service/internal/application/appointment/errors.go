package appointment

import "errors"

var (
	ErrNotFound                             = errors.New("appointment not found")
	ErrInvalidStatus                        = errors.New("invalid appointment status for this operation")
	ErrCannotCancel                         = errors.New("appointment cannot be cancelled")
	ErrUnauthorized                         = errors.New("unauthorized access to appointment")
	ErrInvalidPaymentResultStatus           = errors.New("invalid payment result status")
	ErrPaymentResultConflict                = errors.New("payment result conflicts with current appointment state")
	ErrPaymentEligibilityForbidden          = errors.New("appointment does not belong to payer")
	ErrPaymentEligibilityConflict           = errors.New("appointment is not eligible for payment")
	ErrInvalidBookingPrice                  = errors.New("invalid booking price")
	ErrReadRepositoryUnavailable            = errors.New("appointment read repository unavailable")
	ErrMedicalRecordNotFound                = errors.New("medical record not found")
	ErrMedicalRecordAppointmentNotConfirmed = errors.New("appointment must be confirmed or completed to add medical record")
)
