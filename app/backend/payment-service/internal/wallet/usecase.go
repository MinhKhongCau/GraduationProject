package wallet

import (
	"context"
	"errors"
	"fmt"
	"log"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"
	"payment-service/pkg/redis"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrLockFailed          = errors.New("failed to acquire distributed lock")
	ErrMaxRetriesReached   = errors.New("failed to update wallet after maximum retries due to concurrent updates")
)

type Usecase interface {
	GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	GetWalletByID(ctx context.Context, walletID uuid.UUID) (*entity.Wallet, error)
	GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]entity.WalletTransaction, error)

	CreditPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	CreditAvailable(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	DebitAvailable(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error
	DebitPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error

	LockFunds(ctx context.Context, userID uuid.UUID, amount vo.Money) error
	UnlockFunds(ctx context.Context, userID uuid.UUID, amount vo.Money) error
}

type walletUsecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &walletUsecase{repo: repo}
}

func (u *walletUsecase) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	wallet, err := u.repo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newWallet := &entity.Wallet{
				UserID:           userID,
				AvailableBalance: 0,
				PendingBalance:   0,
				LockedBalance:    0,
				Version:          1,
			}
			err = u.repo.Create(newWallet)
			if err != nil {
				return nil, err
			}
			return newWallet, nil
		}
		return nil, err
	}
	return wallet, nil
}

func (u *walletUsecase) GetWalletByID(ctx context.Context, walletID uuid.UUID) (*entity.Wallet, error) {
	return u.repo.GetByID(walletID)
}

func (u *walletUsecase) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]entity.WalletTransaction, error) {
	wallet, err := u.GetOrCreateWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetTransactionsByWalletID(wallet.ID)
}

func (u *walletUsecase) CreditPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		wallet.PendingBalance = wallet.PendingBalance.Add(amount)

		// Record transaction
		transaction := &entity.WalletTransaction{
			WalletID:       wallet.ID,
			Type:           entity.TxTypePaymentReceived,
			Amount:         amount,
			BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
			ReferenceType:  refType,
			ReferenceID:    refID,
			IdempotencyKey: idempotencyKey,
		}

		return u.repo.CreateTransaction(tx, transaction)
	})
}

func (u *walletUsecase) CreditAvailable(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		wallet.AvailableBalance = wallet.AvailableBalance.Add(amount)

		transaction := &entity.WalletTransaction{
			WalletID:       wallet.ID,
			Type:           entity.TxTypeRefund, // e.g. REFUND / ADJUSTMENT
			Amount:         amount,
			BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
			ReferenceType:  refType,
			ReferenceID:    refID,
			IdempotencyKey: idempotencyKey,
		}

		return u.repo.CreateTransaction(tx, transaction)
	})
}

func (u *walletUsecase) DebitAvailable(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		if wallet.AvailableBalance.Int64() < amount.Int64() {
			return ErrInsufficientBalance
		}
		wallet.AvailableBalance = wallet.AvailableBalance.Sub(amount)

		transaction := &entity.WalletTransaction{
			WalletID:       wallet.ID,
			Type:           entity.TxTypeCommissionDeducted, // e.g. COMMISSION_DEDUCTED
			Amount:         vo.Money(-amount.Int64()),       // Negative amount representation
			BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
			ReferenceType:  refType,
			ReferenceID:    refID,
			IdempotencyKey: idempotencyKey,
		}

		return u.repo.CreateTransaction(tx, transaction)
	})
}

func (u *walletUsecase) DebitPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		if wallet.PendingBalance.Int64() < amount.Int64() {
			return ErrInsufficientBalance
		}
		wallet.PendingBalance = wallet.PendingBalance.Sub(amount)

		transaction := &entity.WalletTransaction{
			WalletID:       wallet.ID,
			Type:           entity.TxTypeCommissionDeducted,
			Amount:         vo.Money(-amount.Int64()),
			BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
			ReferenceType:  refType,
			ReferenceID:    refID,
			IdempotencyKey: idempotencyKey,
		}

		return u.repo.CreateTransaction(tx, transaction)
	})
}

func (u *walletUsecase) LockFunds(ctx context.Context, userID uuid.UUID, amount vo.Money) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		if wallet.AvailableBalance.Int64() < amount.Int64() {
			return ErrInsufficientBalance
		}
		wallet.AvailableBalance = wallet.AvailableBalance.Sub(amount)
		wallet.LockedBalance = wallet.LockedBalance.Add(amount)
		return nil
	})
}

func (u *walletUsecase) UnlockFunds(ctx context.Context, userID uuid.UUID, amount vo.Money) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		if wallet.LockedBalance.Int64() < amount.Int64() {
			return errors.New("cannot unlock more than locked balance")
		}
		wallet.LockedBalance = wallet.LockedBalance.Sub(amount)
		wallet.AvailableBalance = wallet.AvailableBalance.Add(amount)
		return nil
	})
}

// Helper to execute operations under a Redis lock and retry on optimistic locking collision
func (u *walletUsecase) executeWithLockAndRetry(ctx context.Context, userID uuid.UUID, fn func(wallet *entity.Wallet, tx *gorm.DB) error) error {
	// TODO: RabbitMQ and Redis connection logging
	log.Printf("[REDIS LOCK] Attempting to acquire lock for user %s", userID)

	lockKey := fmt.Sprintf("wallet:%s", userID.String())
	acquired, err := redis.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil {
		return err
	}
	if !acquired {
		return ErrLockFailed
	}
	defer func() {
		log.Printf("[REDIS LOCK] Releasing lock for user %s", userID)
		_ = redis.ReleaseLock(ctx, lockKey)
	}()

	maxRetries := 5
	var opErr error

	for i := 0; i < maxRetries; i++ {
		opErr = u.repo.WithTransaction(func(tx *gorm.DB) error {
			// Get or create wallet inside transaction
			var wallet entity.Wallet
			if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					wallet = entity.Wallet{
						UserID:           userID,
						AvailableBalance: 0,
						PendingBalance:   0,
						LockedBalance:    0,
						Version:          1,
					}
					if err := tx.Create(&wallet).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			// Apply business logic
			if err := fn(&wallet, tx); err != nil {
				return err
			}

			// Save wallet using repository optimistic lock check
			return u.repo.UpdateWithTx(tx, &wallet)
		})

		if opErr == nil {
			return nil
		}

		if errors.Is(opErr, ErrConcurrentUpdate) {
			log.Printf("[OPTIMISTIC LOCK RETRY] Retry %d/%d for user %s due to concurrent update", i+1, maxRetries, userID)
			time.Sleep(50 * time.Millisecond) // exponential/random backoff could be used
			continue
		}

		return opErr
	}

	return ErrMaxRetriesReached
}
