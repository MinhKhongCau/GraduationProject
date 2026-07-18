package appointment

import appappointment "booking-service/internal/booking/application/appointment"

type PaymentResultStatus = appappointment.PaymentResultStatus

const (
	PaymentResultSuccess = appappointment.PaymentResultSuccess
	PaymentResultFailed  = appappointment.PaymentResultFailed
)

type HandlePaymentResultCommand = appappointment.HandlePaymentResultCommand

func ParsePaymentResultStatus(status string) (PaymentResultStatus, error) {
	return appappointment.ParsePaymentResultStatus(status)
}
