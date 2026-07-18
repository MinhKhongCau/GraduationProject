package application

import (
	"context"
	"payment-service/internal/booking/client"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"

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

type Repository interface {
	Create(order *entity.PaymentOrder) error
	GetByID(orderID uuid.UUID) (*entity.PaymentOrder, error)
	GetByGatewayTxnRef(ref string) (*entity.PaymentOrder, error)
	Update(order *entity.PaymentOrder) error
}

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	GetOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*entity.PaymentOrder, error)
	GetGatewayTxnRef(ctx context.Context, ref string) (*entity.PaymentOrder, error)
	UpdateOrder(ctx context.Context, order *entity.PaymentOrder) error
	SaveOutboxEvent(ctx context.Context, event *entity.OutboxEvent) error
	CreditWalletPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refID uuid.UUID, idempotencyKey string) error
	DebitWalletPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refID uuid.UUID, idempotencyKey string) error
}

type paymentUsecase struct {
	repo          Repository
	uow           UnitOfWork
	vnpayClient   PaymentGateway
	bookingClient client.BookingServiceClient
}

func NewUsecase(repo Repository, uow UnitOfWork, vnpayClient PaymentGateway, bookingClient client.BookingServiceClient) Usecase {
	return &paymentUsecase{
		repo:          repo,
		uow:           uow,
		vnpayClient:   vnpayClient,
		bookingClient: bookingClient,
	}
}
