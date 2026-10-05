package wallet

import (
	"errors"
	walletdomain "payment-service/internal/domain/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrConcurrentUpdate = errors.New("concurrent wallet update detected")

type Repository interface {
	GetByUserID(userID uuid.UUID) (*walletdomain.Wallet, error)
	GetByID(walletID uuid.UUID) (*walletdomain.Wallet, error)
	Create(wallet *walletdomain.Wallet) error
	Update(wallet *walletdomain.Wallet) error
	UpdateWithTx(tx *gorm.DB, wallet *walletdomain.Wallet) error
	CreateTransaction(tx *gorm.DB, transaction *walletdomain.WalletTransaction) error
	GetTransactionsByWalletID(walletID uuid.UUID) ([]walletdomain.WalletTransaction, error)
	WithTransaction(fn func(tx *gorm.DB) error) error
}
