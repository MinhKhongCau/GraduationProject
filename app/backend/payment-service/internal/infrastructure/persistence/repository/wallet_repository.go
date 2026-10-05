package repository

import (
	appwallet "payment-service/internal/application/wallet"
	walletdomain "payment-service/internal/domain/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) appwallet.Repository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetByUserID(userID uuid.UUID) (*walletdomain.Wallet, error) {
	var wallet walletdomain.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByID(walletID uuid.UUID) (*walletdomain.Wallet, error) {
	var wallet walletdomain.Wallet
	err := r.db.Where("id = ?", walletID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Create(wallet *walletdomain.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) Update(wallet *walletdomain.Wallet) error {
	return r.UpdateWithTx(r.db, wallet)
}

func (r *walletRepository) UpdateWithTx(tx *gorm.DB, wallet *walletdomain.Wallet) error {
	oldVersion := wallet.Version
	wallet.Version++
	result := tx.Model(&walletdomain.Wallet{}).
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
		return appwallet.ErrConcurrentUpdate
	}
	return nil
}

func (r *walletRepository) CreateTransaction(tx *gorm.DB, transaction *walletdomain.WalletTransaction) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(transaction).Error
}

func (r *walletRepository) GetTransactionsByWalletID(walletID uuid.UUID) ([]walletdomain.WalletTransaction, error) {
	var txs []walletdomain.WalletTransaction
	err := r.db.Where("wallet_id = ?", walletID).Order("created_at DESC").Find(&txs).Error
	return txs, err
}

func (r *walletRepository) ListTransactions(walletID uuid.UUID, filter appwallet.TransactionHistoryQuery) ([]walletdomain.WalletTransaction, int64, error) {
	query := r.db.Model(&walletdomain.WalletTransaction{}).Where("wallet_id = ? AND created_at >= ? AND created_at < ?", walletID, filter.FromMs, filter.ToMs)
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Direction == "CREDIT" {
		query = query.Where("amount > 0")
	}
	if filter.Direction == "DEBIT" {
		query = query.Where("amount < 0")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []walletdomain.WalletTransaction
	err := query.Order("created_at DESC, id DESC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&items).Error
	return items, total, err
}

func (r *walletRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
