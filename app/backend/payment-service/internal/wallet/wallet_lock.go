package wallet

import (
	"context"
	"errors"
	"payment-service/internal/domain/entity"
	"payment-service/internal/domain/vo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
