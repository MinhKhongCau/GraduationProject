package payment

import (
	"payment-service/internal/booking/client"
	apppayment "payment-service/internal/payment/application"
	"payment-service/internal/wallet"
)

type Usecase = apppayment.Usecase
type PaymentGateway = apppayment.PaymentGateway
type UnitOfWork = apppayment.UnitOfWork

type unitOfWorkFactory interface {
	NewUnitOfWork(walletUsecase wallet.Usecase) UnitOfWork
}

func NewUsecase(repo Repository, walletUsecase wallet.Usecase, vnpayClient PaymentGateway, bookingClient client.BookingServiceClient) Usecase {
	var uow UnitOfWork
	if factory, ok := repo.(unitOfWorkFactory); ok {
		uow = factory.NewUnitOfWork(walletUsecase)
	}
	return apppayment.NewUsecase(repo, uow, vnpayClient, bookingClient)
}
