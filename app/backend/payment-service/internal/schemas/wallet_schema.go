// File: internal/schemas/wallet_schema.go
package schemas

import "github.com/shopspring/decimal"

// Request tạo ví lần đầu
type InitWalletRequest struct {
	OwnerID  string `json:"owner_id" binding:"required"`
	UserType string `json:"user_type" binding:"required"` // "PATIENT" hoặc "EXPERT"
}

// Request khi Nạp tiền
type TopUpRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

// File: internal/schemas/wallet_schema.go

// Request khi Bệnh nhân thanh toán cho lịch hẹn
type PaymentRequest struct {
	PayerID        string          `json:"payer_id" binding:"required"` // ID Bệnh nhân
	PayeeID        string          `json:"payee_id" binding:"required"` // ID Chuyên gia
	Amount         decimal.Decimal `json:"amount" binding:"required"`
	RelatedOrderID string          `json:"related_order_id" binding:"required"` // ID của Lịch hẹn (Booking)
}

type CreateWithdrawalRequest struct {
	Amount   decimal.Decimal `json:"amount" binding:"required"`
	BankInfo string          `json:"bank_info" binding:"required"` // Ví dụ: "VCB - 0123456789 - NGUYEN VAN A"
}

type ProcessWithdrawalRequest struct {
	Action    string `json:"action" binding:"required"` // Chỉ nhận: "APPROVE" hoặc "REJECT"
	AdminNote string `json:"admin_note"`
}
