package withdrawal

import (
	"context"
	"errors"
	"payment-service/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateBankAccount(account *entity.BankAccount) error
	GetBankAccountByID(id uuid.UUID) (*entity.BankAccount, error)
	GetBankAccountsByUserID(userID uuid.UUID) ([]entity.BankAccount, error)

	CreateWithdrawalRequest(req *entity.WithdrawalRequest) error
	GetWithdrawalRequestByID(id uuid.UUID) (*entity.WithdrawalRequest, error)
	GetWithdrawalRequestsByWalletID(walletID uuid.UUID) ([]entity.WithdrawalRequest, error)
	UpdateWithdrawalRequest(req *entity.WithdrawalRequest) error
	UpdateWithdrawalRequestWithTx(tx *gorm.DB, req *entity.WithdrawalRequest) error

	SaveOutboxEvent(tx *gorm.DB, event *entity.OutboxEvent) error
	WithTransaction(fn func(tx *gorm.DB) error) error
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateBankAccount(account *entity.BankAccount) error {
	return r.db.Create(account).Error
}

func (r *pgRepository) GetBankAccountByID(id uuid.UUID) (*entity.BankAccount, error) {
	var account entity.BankAccount
	err := r.db.Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *pgRepository) GetBankAccountsByUserID(userID uuid.UUID) ([]entity.BankAccount, error) {
	var accounts []entity.BankAccount
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

func (r *pgRepository) CreateWithdrawalRequest(req *entity.WithdrawalRequest) error {
	return r.db.Create(req).Error
}

func (r *pgRepository) GetWithdrawalRequestByID(id uuid.UUID) (*entity.WithdrawalRequest, error) {
	var request entity.WithdrawalRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *pgRepository) GetWithdrawalRequestsByWalletID(walletID uuid.UUID) ([]entity.WithdrawalRequest, error) {
	var requests []entity.WithdrawalRequest
	err := r.db.Where("wallet_id = ?", walletID).Order("requested_at DESC").Find(&requests).Error
	return requests, err
}

type withdrawalReadRow struct {
	entity.WithdrawalRequest `gorm:"embedded"`
	OwnerID                  uuid.UUID `gorm:"column:owner_id"`
}

func (r *pgRepository) ListWithdrawals(ctx context.Context, filter WithdrawalFilter) ([]WithdrawalRecord, int64, error) {
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
	records := make([]WithdrawalRecord, len(rows))
	for i := range rows {
		records[i] = WithdrawalRecord{Request: rows[i].WithdrawalRequest, OwnerID: rows[i].OwnerID}
	}
	return records, total, err
}

func (r *pgRepository) GetWithdrawal(ctx context.Context, requestID uuid.UUID) (*WithdrawalRecord, error) {
	var row withdrawalReadRow
	err := r.db.WithContext(ctx).Table("payment_withdrawal_requests withdrawal").Select("withdrawal.*, wallet.user_id AS owner_id").Joins("JOIN payment_wallets wallet ON wallet.id = withdrawal.wallet_id").Where("withdrawal.id = ?", requestID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWithdrawalNotFound
	}
	if err != nil {
		return nil, err
	}
	return &WithdrawalRecord{Request: row.WithdrawalRequest, OwnerID: row.OwnerID}, nil
}

func (r *pgRepository) UpdateWithdrawalRequest(req *entity.WithdrawalRequest) error {
	return r.db.Save(req).Error
}

func (r *pgRepository) UpdateWithdrawalRequestWithTx(tx *gorm.DB, req *entity.WithdrawalRequest) error {
	return tx.Save(req).Error
}

func (r *pgRepository) SaveOutboxEvent(tx *gorm.DB, event *entity.OutboxEvent) error {
	return tx.Create(event).Error
}

func (r *pgRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
