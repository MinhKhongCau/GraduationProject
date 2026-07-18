package application

import (
	"context"
	"payment-service/internal/booking/client"
	"payment-service/internal/domain/entity"
	"payment-service/internal/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	GetByIDForUpdate(tx *gorm.DB, orderID uuid.UUID) (*entity.PaymentOrder, error)
	GetByGatewayTxnRef(ref string) (*entity.PaymentOrder, error)
	GetByGatewayTxnRefWithTx(tx *gorm.DB, ref string) (*entity.PaymentOrder, error)
	Update(order *entity.PaymentOrder) error
	UpdateWithTx(tx *gorm.DB, order *entity.PaymentOrder) error
	SaveOutboxEvent(tx *gorm.DB, event *entity.OutboxEvent) error
	WithTransaction(fn func(tx *gorm.DB) error) error
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
