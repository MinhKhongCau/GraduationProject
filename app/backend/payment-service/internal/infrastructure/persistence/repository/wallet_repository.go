package repository

import (
	"context"
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
	query := walletTransactionFilterQuery(r.db.Model(&walletdomain.WalletTransaction{}).Where("wallet_id = ?", walletID), "", filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []walletdomain.WalletTransaction
	err := query.Order("created_at DESC, id DESC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&items).Error
	return items, total, err
}

type userWalletTransactionRow struct {
	walletdomain.WalletTransaction `gorm:"embedded"`
	WalletUserID                   uuid.UUID `gorm:"column:wallet_user_id"`
}

// ListTransactionsForUsers trả sổ cái ví của nhiều người dùng (phạm vi chuyên gia Admin quản lý).
func (r *walletRepository) ListTransactionsForUsers(ctx context.Context, userIDs []uuid.UUID, filter appwallet.TransactionHistoryQuery) ([]appwallet.UserWalletTransaction, int64, error) {
	query := r.db.WithContext(ctx).
		Table("payment_wallet_transactions AS tx").
		Joins("JOIN payment_wallets AS wallet ON wallet.id = tx.wallet_id").
		Where("wallet.user_id IN ?", userIDs)
	query = walletTransactionFilterQuery(query, "tx.", filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []userWalletTransactionRow
	err := query.Select("tx.*, wallet.user_id AS wallet_user_id").
		Order("tx.created_at DESC, tx.id DESC").
		Limit(filter.Page.Size).Offset(filter.Page.Offset()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	items := make([]appwallet.UserWalletTransaction, len(rows))
	for i := range rows {
		items[i] = appwallet.UserWalletTransaction{WalletTransaction: rows[i].WalletTransaction, UserID: rows[i].WalletUserID}
	}
	return items, total, nil
}

func walletTransactionFilterQuery(query *gorm.DB, prefix string, filter appwallet.TransactionHistoryQuery) *gorm.DB {
	query = query.Where(prefix+"created_at >= ? AND "+prefix+"created_at < ?", filter.FromMs, filter.ToMs)
	if filter.Type != nil {
		query = query.Where(prefix+"type = ?", *filter.Type)
	}
	if filter.Direction == "CREDIT" {
		query = query.Where(prefix + "amount > 0")
	}
	if filter.Direction == "DEBIT" {
		query = query.Where(prefix + "amount < 0")
	}
	return query
}

func (r *walletRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
