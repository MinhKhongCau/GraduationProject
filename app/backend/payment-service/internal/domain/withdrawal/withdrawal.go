package withdrawal

import (
	"payment-service/internal/domain/money"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BankAccount struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index;column:user_id"`
	BankCode          string    `json:"bank_code" gorm:"type:varchar(50);not null;column:bank_code"`
	AccountNumber     string    `json:"account_number" gorm:"type:varchar(50);not null;column:account_number"`
	AccountHolderName string    `json:"account_holder_name" gorm:"type:varchar(100);not null;column:account_holder_name"`
	Verified          bool      `json:"verified" gorm:"type:boolean;not null;default:false;column:verified"`
	CreatedAt         int64     `json:"created_at" gorm:"type:bigint;not null;column:created_at"`
}

func (BankAccount) TableName() string {
	return "payment_bank_accounts"
}

func (b *BankAccount) BeforeCreate(tx *gorm.DB) error {
	if b.CreatedAt == 0 {
		b.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

type WithdrawalStatus int

const (
	WithdrawalStatusPending         WithdrawalStatus = 1 // PENDING
	WithdrawalStatusPendingApproval WithdrawalStatus = 2 // PENDING_APPROVAL
	WithdrawalStatusApproved        WithdrawalStatus = 3 // APPROVED
	WithdrawalStatusProcessing      WithdrawalStatus = 4 // PROCESSING
	WithdrawalStatusCompleted       WithdrawalStatus = 5 // COMPLETED
	WithdrawalStatusRejected        WithdrawalStatus = 6 // REJECTED
	WithdrawalStatusFailed          WithdrawalStatus = 7 // FAILED
)

func (s WithdrawalStatus) String() string {
	switch s {
	case WithdrawalStatusPending:
		return "PENDING"
	case WithdrawalStatusPendingApproval:
		return "PENDING_APPROVAL"
	case WithdrawalStatusApproved:
		return "APPROVED"
	case WithdrawalStatusProcessing:
		return "PROCESSING"
	case WithdrawalStatusCompleted:
		return "COMPLETED"
	case WithdrawalStatusRejected:
		return "REJECTED"
	case WithdrawalStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

type ApprovalAction int

const (
	ApprovalActionNone    ApprovalAction = 0
	ApprovalActionApprove ApprovalAction = 1 // APPROVE
	ApprovalActionReject  ApprovalAction = 2 // REJECT
)

func (a ApprovalAction) String() string {
	switch a {
	case ApprovalActionApprove:
		return "APPROVE"
	case ApprovalActionReject:
		return "REJECT"
	default:
		return "NONE"
	}
}

type WithdrawalRequest struct {
	ID                     uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WalletID               uuid.UUID        `json:"wallet_id" gorm:"type:uuid;not null;index;column:wallet_id"`
	BankAccountID          uuid.UUID        `json:"bank_account_id" gorm:"type:uuid;not null;column:bank_account_id"`
	Amount                 money.Money      `json:"amount" gorm:"type:bigint;not null;column:amount"`
	Status                 WithdrawalStatus `json:"status" gorm:"type:integer;not null;default:1;column:status"`
	RequiresManualApproval bool             `json:"requires_manual_approval" gorm:"type:boolean;not null;default:false;column:requires_manual_approval"`
	ApproverID             *uuid.UUID       `json:"approver_id" gorm:"type:uuid;column:approver_id"`
	ApprovalAction         ApprovalAction   `json:"approval_action" gorm:"type:integer;not null;default:0;column:approval_action"`
	ApprovalNote           string           `json:"approval_note" gorm:"type:text;column:approval_note"`
	ApprovedAt             *int64           `json:"approved_at" gorm:"type:bigint;column:approved_at"`
	PayoutRef              string           `json:"payout_ref" gorm:"type:varchar(255);column:payout_ref"`
	RequestedAt            int64            `json:"requested_at" gorm:"type:bigint;not null;column:requested_at"`
	ProcessedAt            *int64           `json:"processed_at" gorm:"type:bigint;column:processed_at"`
}

func (WithdrawalRequest) TableName() string {
	return "payment_withdrawal_requests"
}

func (w *WithdrawalRequest) BeforeCreate(tx *gorm.DB) error {
	if w.RequestedAt == 0 {
		w.RequestedAt = time.Now().UnixMilli()
	}
	if w.Status == 0 {
		w.Status = WithdrawalStatusPending
	}
	return nil
}
