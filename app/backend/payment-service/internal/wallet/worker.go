package wallet

import (
	"context"
	"log"
	"payment-service/internal/domain/entity"
	"time"

	"gorm.io/gorm"
)

type Worker struct {
	db         *gorm.DB
	walletRepo Repository
	holdPeriod time.Duration
}

func NewWorker(db *gorm.DB, walletRepo Repository, holdPeriod time.Duration) *Worker {
	return &Worker{
		db:         db,
		walletRepo: walletRepo,
		holdPeriod: holdPeriod,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Poll frequently in development
	defer ticker.Stop()

	log.Println("⏳ Hold Period Release Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Hold Period Release Worker")
			return
		case <-ticker.C:
			w.releasePendingBalances()
		}
	}
}

func (w *Worker) releasePendingBalances() {
	now := time.Now().UnixMilli()
	holdThreshold := now - w.holdPeriod.Milliseconds()

	var orders []entity.PaymentOrder
	// Find SUCCESS, unreleased orders paid before the hold threshold
	err := w.db.Where("status = ? AND released = ? AND paid_at <= ?", entity.OrderStatusSuccess, false, holdThreshold).Find(&orders).Error
	if err != nil {
		log.Printf("Worker release error finding orders: %v", err)
		return
	}

	for _, order := range orders {
		log.Printf("Worker: Releasing hold for order %s, net_amount=%d, expert_id=%s", order.ID, order.NetAmount, order.ExpertID)

		err := w.db.Transaction(func(tx *gorm.DB) error {
			// Find expert wallet
			var wallet entity.Wallet
			if err := tx.Where("user_id = ?", order.ExpertID).First(&wallet).Error; err != nil {
				return err
			}

			// Update balances: move net_amount from pending to available
			wallet.PendingBalance = wallet.PendingBalance.Sub(order.NetAmount)
			wallet.AvailableBalance = wallet.AvailableBalance.Add(order.NetAmount)

			// Record transaction of type ADJUSTMENT
			// Amount is 0 because total balance (Available + Pending + Locked) doesn't change
			transaction := entity.WalletTransaction{
				WalletID:       wallet.ID,
				Type:           entity.TxTypeAdjustment,
				Amount:         0,
				BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
				ReferenceType:  "PAYMENT_ORDER",
				ReferenceID:    order.ID,
				IdempotencyKey: "release_" + order.ID.String(),
			}

			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}

			// Update wallet using optimistic lock check
			oldVersion := wallet.Version
			wallet.Version++
			result := tx.Model(&entity.Wallet{}).
				Where("id = ? AND version = ?", wallet.ID, oldVersion).
				Updates(map[string]interface{}{
					"available_balance": wallet.AvailableBalance,
					"pending_balance":   wallet.PendingBalance,
					"version":           wallet.Version,
				})

			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrConcurrentUpdate
			}

			// Mark order as released
			order.Released = true
			if err := tx.Save(&order).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Printf("❌ Worker failed to release hold for order %s: %v", order.ID, err)
		} else {
			log.Printf("✅ Worker successfully released hold for order %s", order.ID)
		}
	}
}
