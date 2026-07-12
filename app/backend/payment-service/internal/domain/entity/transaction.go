package entity

import (
	"payment-service/internal/domain/vo"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType int

const (
	TxTypePaymentReceived     TransactionType = 1 // PAYMENT_RECEIVED
	TxTypeCommissionDeducted  TransactionType = 2 // COMMISSION_DEDUCTED
	TxTypeRefund              TransactionType = 3 // REFUND
	TxTypeWithdrawalLocked    TransactionType = 4 // WITHDRAWAL_LOCKED
	TxTypeWithdrawalCompleted TransactionType = 5 // WITHDRAWAL_COMPLETED
	TxTypeWithdrawalRejected  TransactionType = 6 // WITHDRAWAL_REJECTED
	TxTypeAdjustment          TransactionType = 7 // ADJUSTMENT
)

func (t TransactionType) String() string {
	switch t {
	case TxTypePaymentReceived:
		return "PAYMENT_RECEIVED"
	case TxTypeCommissionDeducted:
		return "COMMISSION_DEDUCTED"
	case TxTypeRefund:
		return "REFUND"
	case TxTypeWithdrawalLocked:
		return "WITHDRAWAL_LOCKED"
	case TxTypeWithdrawalCompleted:
		return "WITHDRAWAL_COMPLETED"
	case TxTypeWithdrawalRejected:
		return "WITHDRAWAL_REJECTED"
	case TxTypeAdjustment:
		return "ADJUSTMENT"
	default:
		return "UNKNOWN"
	}
}

type WalletTransaction struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WalletID       uuid.UUID       `json:"wallet_id" gorm:"type:uuid;not null;index;column:wallet_id"`
	Type           TransactionType `json:"type" gorm:"type:integer;not null;column:type"`
	Amount         vo.Money        `json:"amount" gorm:"type:bigint;not null;column:amount"`
	BalanceAfter   vo.Money        `json:"balance_after" gorm:"type:bigint;not null;column:balance_after"`
	ReferenceType  string          `json:"reference_type" gorm:"type:varchar(50);column:reference_type"` // PAYMENT_ORDER | WITHDRAWAL_REQUEST | MANUAL
	ReferenceID    uuid.UUID       `json:"reference_id" gorm:"type:uuid;column:reference_id"`
	IdempotencyKey string          `json:"idempotency_key" gorm:"type:varchar(255);uniqueIndex;column:idempotency_key"`
	CreatedAt      int64           `json:"created_at" gorm:"type:bigint;not null;column:created_at"`
}

func (WalletTransaction) TableName() string {
	return "payment_wallet_transactions"
}

func (t *WalletTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.CreatedAt == 0 {
		t.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}
