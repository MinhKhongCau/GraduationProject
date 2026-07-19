package application

import (
	"context"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	"time"

	"github.com/google/uuid"
)

type Usecase interface {
	// CreateOrder creates a VNPay payment order for a booking appointment.
	CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, ipAddr string) (*entity.PaymentOrder, string, error)
	ProcessIPN(ctx context.Context, params map[string][]string) (bool, error)
	ListCompensationCases(ctx context.Context, filter CompensationCaseFilter) (*CompensationCasePage, error)
	GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*CompensationCase, error)
}

type PaymentGateway interface {
	GeneratePaymentURL(txnRef string, amount int64, ipAddr, desc string, createdAt, expiresAt int64) string
	VerifyChecksum(params map[string][]string) bool
}

type Repository interface {
	Create(order *entity.PaymentOrder) error
	GetByID(orderID uuid.UUID) (*entity.PaymentOrder, error)
	GetByGatewayTxnRef(ref string) (*entity.PaymentOrder, error)
	GetSuccessfulByAppointment(ctx context.Context, appointmentID uuid.UUID) (*entity.PaymentOrder, error)
	Update(order *entity.PaymentOrder) error
}

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	GetOrder(ctx context.Context, orderID uuid.UUID) (*entity.PaymentOrder, error)
	GetOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*entity.PaymentOrder, error)
	ListOrdersForAppointmentForUpdate(ctx context.Context, appointmentID uuid.UUID) ([]entity.PaymentOrder, error)
	CreateOrder(ctx context.Context, order *entity.PaymentOrder) error
	GetGatewayTxnRef(ctx context.Context, ref string) (*entity.PaymentOrder, error)
	UpdateOrder(ctx context.Context, order *entity.PaymentOrder) error
	ExpireOtherPendingOrders(ctx context.Context, appointmentID, exceptOrderID uuid.UUID) error
	SaveOutboxEvent(ctx context.Context, event *entity.OutboxEvent) error
	CreditWalletPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refID uuid.UUID, idempotencyKey string) error
	DebitWalletPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refID uuid.UUID, idempotencyKey string) error
	SaveCompensationCase(ctx context.Context, compensationCase *entity.PaymentCompensationCase) error
}

type paymentUsecase struct {
	repo          Repository
	uow           UnitOfWork
	vnpayClient   PaymentGateway
	bookingClient BookingServiceClient
	orderTTL      time.Duration
	minimumWindow time.Duration
	clock         func() time.Time
}

func NewUsecase(repo Repository, uow UnitOfWork, vnpayClient PaymentGateway, bookingClient BookingServiceClient) Usecase {
	return NewUsecaseWithOptions(repo, uow, vnpayClient, bookingClient, Options{})
}

type Options struct {
	OrderTTL      time.Duration
	MinimumWindow time.Duration
	Clock         func() time.Time
}

func NewUsecaseWithOptions(repo Repository, uow UnitOfWork, vnpayClient PaymentGateway, bookingClient BookingServiceClient, options Options) Usecase {
	if options.OrderTTL <= 0 {
		options.OrderTTL = 15 * time.Minute
	}
	if options.MinimumWindow <= 0 {
		options.MinimumWindow = time.Minute
	}
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &paymentUsecase{
		repo:          repo,
		uow:           uow,
		vnpayClient:   vnpayClient,
		bookingClient: bookingClient,
		orderTTL:      options.OrderTTL,
		minimumWindow: options.MinimumWindow,
		clock:         options.Clock,
	}
}
