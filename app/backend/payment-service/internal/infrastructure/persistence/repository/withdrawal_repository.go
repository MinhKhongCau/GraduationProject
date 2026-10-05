package repository

import (
	"context"
	"errors"
	appwithdrawal "payment-service/internal/application/withdrawal"
	paymentdomain "payment-service/internal/domain/payment"
	withdrawaldomain "payment-service/internal/domain/withdrawal"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type withdrawalRepository struct {
	db *gorm.DB
}

func NewWithdrawalRepository(db *gorm.DB) appwithdrawal.Repository {
	return &withdrawalRepository{db: db}
}

func (r *withdrawalRepository) CreateBankAccount(account *withdrawaldomain.BankAccount) error {
	return r.db.Create(account).Error
}

func (r *withdrawalRepository) GetBankAccountByID(id uuid.UUID) (*withdrawaldomain.BankAccount, error) {
	var account withdrawaldomain.BankAccount
	err := r.db.Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *withdrawalRepository) GetBankAccountsByUserID(userID uuid.UUID) ([]withdrawaldomain.BankAccount, error) {
	var accounts []withdrawaldomain.BankAccount
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

func (r *withdrawalRepository) CreateWithdrawalRequest(req *withdrawaldomain.WithdrawalRequest) error {
	return r.db.Create(req).Error
}

func (r *withdrawalRepository) GetWithdrawalRequestByID(id uuid.UUID) (*withdrawaldomain.WithdrawalRequest, error) {
	var request withdrawaldomain.WithdrawalRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *withdrawalRepository) GetWithdrawalRequestsByWalletID(walletID uuid.UUID) ([]withdrawaldomain.WithdrawalRequest, error) {
	var requests []withdrawaldomain.WithdrawalRequest
	err := r.db.Where("wallet_id = ?", walletID).Order("requested_at DESC").Find(&requests).Error
	return requests, err
}

type withdrawalReadRow struct {
	withdrawaldomain.WithdrawalRequest `gorm:"embedded"`
	OwnerID                            uuid.UUID `gorm:"column:owner_id"`
}

func (r *withdrawalRepository) ListWithdrawals(ctx context.Context, filter appwithdrawal.WithdrawalFilter) ([]appwithdrawal.WithdrawalRecord, int64, error) {
	query := r.db.WithContext(ctx).Table("payment_withdrawal_requests withdrawal").Joins("JOIN payment_wallets wallet ON wallet.id = withdrawal.wallet_id")
	if filter.IsAdmin {
		if filter.ExpertID != nil {
			query = query.Where("wallet.user_id = ?", *filter.ExpertID)
		}
	} else {
		query = query.Where("wallet.user_id = ?", filter.ActorID)
	}
	if filter.Status != nil {
		query = query.Where("withdrawal.status = ?", *filter.Status)
	}
	query = query.Where("withdrawal.requested_at >= ? AND withdrawal.requested_at < ?", filter.FromMs, filter.ToMs)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []withdrawalReadRow
	err := query.Select("withdrawal.*, wallet.user_id AS owner_id").Order("withdrawal.requested_at DESC, withdrawal.id DESC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Scan(&rows).Error
	records := make([]appwithdrawal.WithdrawalRecord, len(rows))
	for i := range rows {
		records[i] = appwithdrawal.WithdrawalRecord{Request: rows[i].WithdrawalRequest, OwnerID: rows[i].OwnerID}
	}
	return records, total, err
}

func (r *withdrawalRepository) GetWithdrawal(ctx context.Context, requestID uuid.UUID) (*appwithdrawal.WithdrawalRecord, error) {
	var row withdrawalReadRow
	err := r.db.WithContext(ctx).Table("payment_withdrawal_requests withdrawal").Select("withdrawal.*, wallet.user_id AS owner_id").Joins("JOIN payment_wallets wallet ON wallet.id = withdrawal.wallet_id").Where("withdrawal.id = ?", requestID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appwithdrawal.ErrWithdrawalNotFound
	}
	if err != nil {
		return nil, err
	}
	return &appwithdrawal.WithdrawalRecord{Request: row.WithdrawalRequest, OwnerID: row.OwnerID}, nil
}

func (r *withdrawalRepository) UpdateWithdrawalRequest(req *withdrawaldomain.WithdrawalRequest) error {
	return r.db.Save(req).Error
}

func (r *withdrawalRepository) UpdateWithdrawalRequestWithTx(tx *gorm.DB, req *withdrawaldomain.WithdrawalRequest) error {
	return tx.Save(req).Error
}

func (r *withdrawalRepository) SaveOutboxEvent(tx *gorm.DB, event *paymentdomain.OutboxEvent) error {
	return tx.Create(event).Error
}

func (r *withdrawalRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
