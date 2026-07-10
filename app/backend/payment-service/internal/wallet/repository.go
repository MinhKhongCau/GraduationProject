package wallet

import (
	"errors"
	"payment-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrConcurrentUpdate = errors.New("concurrent wallet update detected")

type Repository interface {
	GetByUserID(userID uuid.UUID) (*domain.Wallet, error)
	GetByID(walletID uuid.UUID) (*domain.Wallet, error)
	Create(wallet *domain.Wallet) error
	Update(wallet *domain.Wallet) error
	UpdateWithTx(tx *gorm.DB, wallet *domain.Wallet) error
	CreateTransaction(tx *gorm.DB, transaction *domain.WalletTransaction) error
	GetTransactionsByWalletID(walletID uuid.UUID) ([]domain.WalletTransaction, error)
	WithTransaction(fn func(tx *gorm.DB) error) error
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) GetByUserID(userID uuid.UUID) (*domain.Wallet, error) {
	var wallet domain.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *pgRepository) GetByID(walletID uuid.UUID) (*domain.Wallet, error) {
	var wallet domain.Wallet
	err := r.db.Where("id = ?", walletID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *pgRepository) Create(wallet *domain.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *pgRepository) Update(wallet *domain.Wallet) error {
	return r.UpdateWithTx(r.db, wallet)
}

func (r *pgRepository) UpdateWithTx(tx *gorm.DB, wallet *domain.Wallet) error {
	oldVersion := wallet.Version
	wallet.Version++
	result := tx.Model(&domain.Wallet{}).
		Where("id = ? AND version = ?", wallet.ID, oldVersion).
		Updates(map[string]interface{}{
			"available_balance": wallet.AvailableBalance,
			"pending_balance":   wallet.PendingBalance,
			"locked_balance":    wallet.LockedBalance,
			"version":           wallet.Version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConcurrentUpdate
	}
	return nil
}

func (r *pgRepository) CreateTransaction(tx *gorm.DB, transaction *domain.WalletTransaction) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(transaction).Error
}

func (r *pgRepository) GetTransactionsByWalletID(walletID uuid.UUID) ([]domain.WalletTransaction, error) {
	var txs []domain.WalletTransaction
	err := r.db.Where("wallet_id = ?", walletID).Order("created_at DESC").Find(&txs).Error
	return txs, err
}

func (r *pgRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
