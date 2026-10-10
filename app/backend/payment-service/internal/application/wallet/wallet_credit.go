package wallet

import (
	"context"
	"payment-service/internal/domain/money"
	walletdomain "payment-service/internal/domain/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (u *walletUsecase) CreditPending(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *walletdomain.Wallet, tx *gorm.DB) error {
		return u.applyCreditPending(tx, wallet, amount, refType, refID, idempotencyKey)
	})
}

func (u *walletUsecase) CreditPendingWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithExistingTx(ctx, tx, userID, func(wallet *walletdomain.Wallet, tx *gorm.DB) error {
		return u.applyCreditPending(tx, wallet, amount, refType, refID, idempotencyKey)
	})
}

func (u *walletUsecase) CreditAvailable(ctx context.Context, userID uuid.UUID, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithLockAndRetry(ctx, userID, func(wallet *walletdomain.Wallet, tx *gorm.DB) error {
		wallet.AvailableBalance = wallet.AvailableBalance.Add(amount)

		transaction := &walletdomain.WalletTransaction{
			WalletID:       wallet.ID,
			Type:           walletdomain.TxTypeRefund, // e.g. REFUND / ADJUSTMENT
			Amount:         amount,
			BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
			ReferenceType:  refType,
			ReferenceID:    refID,
			IdempotencyKey: idempotencyKey,
		}

		return u.repo.CreateTransaction(tx, transaction)
	})
}

func (u *walletUsecase) applyCreditPending(tx *gorm.DB, wallet *walletdomain.Wallet, amount money.Money, refType string, refID uuid.UUID, idempotencyKey string) error {
	wallet.PendingBalance = wallet.PendingBalance.Add(amount)

	transaction := &walletdomain.WalletTransaction{
		WalletID:       wallet.ID,
		Type:           walletdomain.TxTypePaymentReceived,
		Amount:         amount,
		BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
		ReferenceType:  refType,
		ReferenceID:    refID,
		IdempotencyKey: idempotencyKey,
	}

	return u.repo.CreateTransaction(tx, transaction)
}

func (u *walletUsecase) AdjustWithTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID, availableDelta, pendingDelta money.Money, txType walletdomain.TransactionType, refType string, refID uuid.UUID, idempotencyKey string) error {
	return u.executeWithExistingTx(ctx, tx, userID, func(wallet *walletdomain.Wallet, tx *gorm.DB) error {
		available := wallet.AvailableBalance.Add(availableDelta)
		pending := wallet.PendingBalance.Add(pendingDelta)
		if available.IsNegative() || pending.IsNegative() {
			return ErrInsufficientBalance
		}
		wallet.AvailableBalance = available
		wallet.PendingBalance = pending

		transaction := &walletdomain.WalletTransaction{
			WalletID:       wallet.ID,
			Type:           txType,
			Amount:         availableDelta.Add(pendingDelta),
			BalanceAfter:   wallet.AvailableBalance.Add(wallet.PendingBalance).Add(wallet.LockedBalance),
			ReferenceType:  refType,
			ReferenceID:    refID,
			IdempotencyKey: idempotencyKey,
		}

		return u.repo.CreateTransaction(tx, transaction)
	})
}
