package withdrawal

import (
	paymentdomain "payment-service/internal/domain/payment"
	withdrawaldomain "payment-service/internal/domain/withdrawal"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateBankAccount(account *withdrawaldomain.BankAccount) error
	GetBankAccountByID(id uuid.UUID) (*withdrawaldomain.BankAccount, error)
	GetBankAccountsByUserID(userID uuid.UUID) ([]withdrawaldomain.BankAccount, error)

	CreateWithdrawalRequest(req *withdrawaldomain.WithdrawalRequest) error
	GetWithdrawalRequestByID(id uuid.UUID) (*withdrawaldomain.WithdrawalRequest, error)
	GetWithdrawalRequestsByWalletID(walletID uuid.UUID) ([]withdrawaldomain.WithdrawalRequest, error)
	UpdateWithdrawalRequest(req *withdrawaldomain.WithdrawalRequest) error
	UpdateWithdrawalRequestWithTx(tx *gorm.DB, req *withdrawaldomain.WithdrawalRequest) error

	SaveOutboxEvent(tx *gorm.DB, event *paymentdomain.OutboxEvent) error
	WithTransaction(fn func(tx *gorm.DB) error) error
}
