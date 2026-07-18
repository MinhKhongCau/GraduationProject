package appointment

import appappointment "booking-service/internal/booking/application/appointment"

var (
	ErrNotFound                    = appappointment.ErrNotFound
	ErrInvalidStatus               = appappointment.ErrInvalidStatus
	ErrCannotCancel                = appappointment.ErrCannotCancel
	ErrUnauthorized                = appappointment.ErrUnauthorized
	ErrInvalidPaymentResultStatus  = appappointment.ErrInvalidPaymentResultStatus
	ErrPaymentResultConflict       = appappointment.ErrPaymentResultConflict
	ErrPaymentEligibilityForbidden = appappointment.ErrPaymentEligibilityForbidden
	ErrPaymentEligibilityConflict  = appappointment.ErrPaymentEligibilityConflict
	ErrInvalidBookingPrice         = appappointment.ErrInvalidBookingPrice
)
