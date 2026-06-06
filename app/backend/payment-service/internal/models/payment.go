// File: internal/models/payment.go
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ==========================================
// 1. VÍ ĐIỆN TỬ (WALLET)
// ==========================================
type Wallet struct {
	WalletID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OwnerID  uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"` // Có thể là PatientID hoặc ExpertID
	UserType string    `gorm:"type:varchar(50);not null"`      // VD: "PATIENT", "EXPERT"

	// Balance dùng decimal để chính xác từng đồng. Mặc định ví mới có 0đ.
	Balance decimal.Decimal `gorm:"type:decimal(12,2);default:0.00"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Quan hệ 1-N: 1 Ví có nhiều Giao dịch và Yêu cầu rút tiền
	Transactions       []Transaction       `gorm:"foreignKey:WalletID"`
	WithdrawalRequests []WithdrawalRequest `gorm:"foreignKey:WalletID"`
}

// ==========================================
// 2. LỊCH SỬ GIAO DỊCH (TRANSACTION)
// ==========================================
type Transaction struct {
	TxnID    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WalletID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Composite Unique Index: Ngăn chặn trừ tiền 2 lần cho cùng 1 đơn hàng và cùng 1 hành động
	RelatedOrderID  uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_unique_order_transaction"`
	TransactionType string    `gorm:"type:varchar(50);uniqueIndex:idx_unique_order_transaction"` // TOP_UP, PAYMENT, REFUND, WITHDRAW

	Amount        decimal.Decimal `gorm:"type:decimal(12,2);not null"`
	BalanceBefore decimal.Decimal `gorm:"type:decimal(12,2);not null"` // Số dư trước khi biến động
	BalanceAfter  decimal.Decimal `gorm:"type:decimal(12,2);not null"` // Số dư sau khi biến động

	Status    string    `gorm:"type:varchar(50);default:'PENDING'"` // PENDING, SUCCESS, FAILED
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// ==========================================
// 3. YÊU CẦU RÚT TIỀN (WITHDRAWAL REQUEST)
// ==========================================
type WithdrawalRequest struct {
	RequestID uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WalletID  uuid.UUID       `gorm:"type:uuid;not null;index"`
	Amount    decimal.Decimal `gorm:"type:decimal(12,2);not null"`
	BankInfo  string          `gorm:"type:text;not null"`
	Status    string          `gorm:"type:varchar(50);default:'PENDING'"` // PENDING, APPROVED, REJECTED
	AdminNote string          `gorm:"type:text"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
}
