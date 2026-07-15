package withdrawal

import (
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
