package wallet

import (
	"context"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (u *walletUsecase) CreditPending(ctx context.Context, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		return u.applyCreditPending(tx, wallet, amount, refType, refID, idempotencyKey)
	})
}

func (u *walletUsecase) CreditPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithExistingTx(ctx, tx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		return u.applyCreditPending(tx, wallet, amount, refType, refID, idempotencyKey)
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

func (u *walletUsecase) applyCreditPending(tx *gorm.DB, wallet *entity.Wallet, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	wallet.PendingBalance = wallet.PendingBalance.Add(amount)

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
}
