package wallet

import (
	"context"
	"errors"
	"payment-service/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
