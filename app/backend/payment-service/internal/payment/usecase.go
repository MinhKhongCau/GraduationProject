package payment

import (
	"payment-service/internal/booking/client"
	apppayment "payment-service/internal/payment/application"
	"payment-service/internal/wallet"
)

type Usecase = apppayment.Usecase
type PaymentGateway = apppayment.PaymentGateway

func NewUsecase(repo Repository, walletUsecase wallet.Usecase, vnpayClient PaymentGateway, bookingClient client.BookingServiceClient) Usecase {
	return apppayment.NewUsecase(repo, walletUsecase, vnpayClient, bookingClient)
}
