package appointment

import "strings"

type PaymentResultStatus string

const (
	PaymentResultSuccess PaymentResultStatus = "SUCCESS"
	PaymentResultFailed  PaymentResultStatus = "FAILED"
)

type HandlePaymentResultCommand struct {
	AppointmentID string
	Status        PaymentResultStatus
}

func ParsePaymentResultStatus(status string) (PaymentResultStatus, error) {
	switch PaymentResultStatus(strings.TrimSpace(status)) {
	case PaymentResultSuccess:
		return PaymentResultSuccess, nil
	case PaymentResultFailed:
		return PaymentResultFailed, nil
	default:
		return "", ErrInvalidPaymentResultStatus
	}
}

func (u *appointmentUsecase) ConfirmPayment(appointmentID string) error {
	return u.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultSuccess,
	})
}

func (u *appointmentUsecase) HandlePaymentFailure(appointmentID string) error {
	return u.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultFailed,
	})
}

func (u *appointmentUsecase) HandlePaymentResult(command HandlePaymentResultCommand) error {
	if command.AppointmentID == "" {
		return ErrNotFound
	}
	if command.Status != PaymentResultSuccess && command.Status != PaymentResultFailed {
		return ErrInvalidPaymentResultStatus
	}
	return u.repo.HandlePaymentResult(command)
}
