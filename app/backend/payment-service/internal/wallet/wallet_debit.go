package wallet

import (
	"context"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
		return u.applyDebitPending(tx, wallet, amount, refType, refID, idempotencyKey)
	})
}

func (u *walletUsecase) DebitPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithExistingTx(ctx, tx, userID, func(wallet *entity.Wallet, tx *gorm.DB) error {
		return u.applyDebitPending(tx, wallet, amount, refType, refID, idempotencyKey)
	})
}

func (u *walletUsecase) applyDebitPending(tx *gorm.DB, wallet *entity.Wallet, amount vo.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
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
}
