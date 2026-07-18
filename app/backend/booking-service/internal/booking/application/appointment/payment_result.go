package appointment

type PaymentResultStatus string

const (
	PaymentResultSuccess PaymentResultStatus = "SUCCESS"
	PaymentResultFailed  PaymentResultStatus = "FAILED"
)

type HandlePaymentResultCommand struct {
	AppointmentID string
	Status        PaymentResultStatus
}
