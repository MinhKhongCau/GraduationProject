package appointmentpostgres

import appappointment "booking-service/internal/booking/application/appointment"

type Repository = appappointment.Repository
type Tx = appappointment.Tx
type HandlePaymentResultCommand = appappointment.HandlePaymentResultCommand
type GetPaymentEligibilityCommand = appappointment.GetPaymentEligibilityCommand
type PaymentEligibilitySnapshot = appappointment.PaymentEligibilitySnapshot

var (
	ErrNotFound                   = appappointment.ErrNotFound
	ErrPaymentResultConflict      = appappointment.ErrPaymentResultConflict
	ErrPaymentEligibilityConflict = appappointment.ErrPaymentEligibilityConflict
	errTxRecordNotFound           = appappointment.ErrTxRecordNotFound
)

const (
	PaymentResultSuccess = appappointment.PaymentResultSuccess
	PaymentResultFailed  = appappointment.PaymentResultFailed
)
