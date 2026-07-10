package withdrawal

import (
	"payment-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateBankAccount(account *domain.BankAccount) error
	GetBankAccountByID(id uuid.UUID) (*domain.BankAccount, error)
	GetBankAccountsByUserID(userID uuid.UUID) ([]domain.BankAccount, error)

	CreateWithdrawalRequest(req *domain.WithdrawalRequest) error
	GetWithdrawalRequestByID(id uuid.UUID) (*domain.WithdrawalRequest, error)
	GetWithdrawalRequestsByWalletID(walletID uuid.UUID) ([]domain.WithdrawalRequest, error)
	UpdateWithdrawalRequest(req *domain.WithdrawalRequest) error
	UpdateWithdrawalRequestWithTx(tx *gorm.DB, req *domain.WithdrawalRequest) error

	SaveOutboxEvent(tx *gorm.DB, event *domain.OutboxEvent) error
	WithTransaction(fn func(tx *gorm.DB) error) error
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateBankAccount(account *domain.BankAccount) error {
	return r.db.Create(account).Error
}

func (r *pgRepository) GetBankAccountByID(id uuid.UUID) (*domain.BankAccount, error) {
	var account domain.BankAccount
	err := r.db.Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *pgRepository) GetBankAccountsByUserID(userID uuid.UUID) ([]domain.BankAccount, error) {
	var accounts []domain.BankAccount
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

func (r *pgRepository) CreateWithdrawalRequest(req *domain.WithdrawalRequest) error {
	return r.db.Create(req).Error
}

func (r *pgRepository) GetWithdrawalRequestByID(id uuid.UUID) (*domain.WithdrawalRequest, error) {
	var request domain.WithdrawalRequest
	err := r.db.Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *pgRepository) GetWithdrawalRequestsByWalletID(walletID uuid.UUID) ([]domain.WithdrawalRequest, error) {
	var requests []domain.WithdrawalRequest
	err := r.db.Where("wallet_id = ?", walletID).Order("requested_at DESC").Find(&requests).Error
	return requests, err
}

func (r *pgRepository) UpdateWithdrawalRequest(req *domain.WithdrawalRequest) error {
	return r.db.Save(req).Error
}

func (r *pgRepository) UpdateWithdrawalRequestWithTx(tx *gorm.DB, req *domain.WithdrawalRequest) error {
	return tx.Save(req).Error
}

func (r *pgRepository) SaveOutboxEvent(tx *gorm.DB, event *domain.OutboxEvent) error {
	return tx.Create(event).Error
}

func (r *pgRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
