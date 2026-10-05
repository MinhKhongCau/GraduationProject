package wallet

import (
	"context"
	"errors"
	"payment-service/internal/application/readquery"
	"payment-service/internal/domain/money"
	walletdomain "payment-service/internal/domain/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrLockFailed          = errors.New("failed to acquire distributed lock")
	ErrMaxRetriesReached   = errors.New("failed to update wallet after maximum retries due to concurrent updates")
)

type Usecase interface {
	GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*walletdomain.Wallet, error)
	GetWalletByID(ctx context.Context, walletID uuid.UUID) (*walletdomain.Wallet, error)
	GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]walletdomain.WalletTransaction, error)

	CreditPending(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	CreditPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	CreditAvailable(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	DebitAvailable(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	DebitPending(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	DebitPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error

	LockFunds(ctx context.Context, userID uuid.UUID, amount money.Money) error
	UnlockFunds(ctx context.Context, userID uuid.UUID, amount money.Money) error
}

type ReadUsecase interface {
	ListTransactionHistory(ctx context.Context, query TransactionHistoryQuery) (*readquery.Page[walletdomain.WalletTransaction], error)
}

type walletUsecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &walletUsecase{repo: repo}
}
