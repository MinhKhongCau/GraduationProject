package payment

import (
	"context"
	"payment-service/internal/booking/client"
	"payment-service/internal/domain/entity"
	"payment-service/internal/wallet"

	"github.com/google/uuid"
)

type Usecase interface {
	// CreateOrder creates a VNPay payment order for a booking appointment.
	CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, ipAddr string) (*entity.PaymentOrder, string, error)
	ProcessIPN(ctx context.Context, params map[string][]string) (bool, error)
}

type PaymentGateway interface {
	GeneratePaymentURL(txnRef string, amount int64, ipAddr, desc, createDate string) string
	VerifyChecksum(params map[string][]string) bool
}

type paymentUsecase struct {
	repo          Repository
	walletUsecase wallet.Usecase
	vnpayClient   PaymentGateway
	bookingClient client.BookingServiceClient
}

func NewUsecase(repo Repository, walletUsecase wallet.Usecase, vnpayClient PaymentGateway, bookingClient client.BookingServiceClient) Usecase {
	return &paymentUsecase{
		repo:          repo,
		walletUsecase: walletUsecase,
		vnpayClient:   vnpayClient,
		bookingClient: bookingClient,
	}
}
