package payment

import (
	"context"
	"payment-service/internal/application/readquery"
	"payment-service/internal/domain/money"
	paymentdomain "payment-service/internal/domain/payment"
	"time"

	"github.com/google/uuid"
)

type Usecase interface {
	// CreateOrder creates a VNPay payment order for a booking appointment.
	CreateOrder(ctx context.Context, payerID uuid.UUID, appointmentID string, ipAddr string) (*paymentdomain.PaymentOrder, string, error)
	ProcessIPN(ctx context.Context, params map[string][]string) (bool, error)
	ListCompensationCases(ctx context.Context, filter CompensationCaseFilter) (*CompensationCasePage, error)
	GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*CompensationCase, error)
}

type ReadUsecase interface {
	ListPaymentOrders(ctx context.Context, filter PaymentOrderFilter) (*readquery.Page[PaymentOrderView], error)
	GetPaymentOrder(ctx context.Context, payerID, orderID uuid.UUID, isAdmin bool) (*PaymentOrderView, error)
}

type PaymentGateway interface {
	GeneratePaymentURL(txnRef string, amount int64, ipAddr, desc string, createdAt, expiresAt int64) string
	VerifyChecksum(params map[string][]string) bool
}

type Repository interface {
	Create(order *paymentdomain.PaymentOrder) error
	GetByID(orderID uuid.UUID) (*paymentdomain.PaymentOrder, error)
	GetByGatewayTxnRef(ref string) (*paymentdomain.PaymentOrder, error)
	GetSuccessfulByAppointment(ctx context.Context, appointmentID uuid.UUID) (*paymentdomain.PaymentOrder, error)
	Update(order *paymentdomain.PaymentOrder) error
}

type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	GetOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error)
	GetOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error)
	ListOrdersForAppointmentForUpdate(ctx context.Context, appointmentID uuid.UUID) ([]paymentdomain.PaymentOrder, error)
	CreateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error
	GetGatewayTxnRef(ctx context.Context, ref string) (*paymentdomain.PaymentOrder, error)
	UpdateOrder(ctx context.Context, order *paymentdomain.PaymentOrder) error
	ExpireOtherPendingOrders(ctx context.Context, appointmentID, exceptOrderID uuid.UUID) error
	SaveOutboxEvent(ctx context.Context, event *paymentdomain.OutboxEvent) error
	CreditWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error
	DebitWalletPending(ctx context.Context, userID uuid.UUID, amount money.Money, refID uuid.UUID, idempotencyKey string) error
	SaveCompensationCase(ctx context.Context, compensationCase *paymentdomain.PaymentCompensationCase) error
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
