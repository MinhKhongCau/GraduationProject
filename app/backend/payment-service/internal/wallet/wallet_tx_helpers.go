package wallet

import (
	"context"
	"errors"
	"fmt"
	"log"
	"payment-service/internal/domain/entity"
	"payment-service/pkg/redis"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

func (u *walletUsecase) executeWithExistingTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, fn func(wallet *entity.Wallet, tx *gorm.DB) error) error {
	if tx == nil {
		return errors.New("wallet transaction is required")
	}

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

	var wallet entity.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
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

	if err := fn(&wallet, tx); err != nil {
		return err
	}

	return u.repo.UpdateWithTx(tx, &wallet)
}
