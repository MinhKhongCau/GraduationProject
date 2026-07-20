package wallet

import (
	"context"
	"errors"
	"fmt"
	"payment-service/internal/domain/entity"
	"payment-service/internal/payment/application/readquery"
	"strings"

	"github.com/google/uuid"
)

type TransactionHistoryQuery struct {
	UserID    uuid.UUID
	Type      *entity.TransactionType
	Direction string
	FromMs    int64
	ToMs      int64
	Page      readquery.PageRequest
}
type transactionReader interface {
	ListTransactions(walletID uuid.UUID, query TransactionHistoryQuery) ([]entity.WalletTransaction, int64, error)
}

func (u *walletUsecase) ListTransactionHistory(ctx context.Context, filter TransactionHistoryQuery) (*readquery.Page[entity.WalletTransaction], error) {
	wallet, err := u.GetOrCreateWallet(ctx, filter.UserID)
	if err != nil {
		return nil, err
	}
	reader, ok := u.repo.(transactionReader)
	if !ok {
		return nil, errors.New("wallet transaction reader unavailable")
	}
	items, total, err := reader.ListTransactions(wallet.ID, filter)
	if err != nil {
		return nil, err
	}
	page := readquery.NewPage(items, filter.Page, total)
	return &page, nil
}

func ParseTransactionType(value string) (*entity.TransactionType, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}
	values := map[string]entity.TransactionType{"PAYMENT_RECEIVED": entity.TxTypePaymentReceived, "COMMISSION_DEDUCTED": entity.TxTypeCommissionDeducted, "REFUND": entity.TxTypeRefund, "WITHDRAWAL_LOCKED": entity.TxTypeWithdrawalLocked, "WITHDRAWAL_COMPLETED": entity.TxTypeWithdrawalCompleted, "WITHDRAWAL_REJECTED": entity.TxTypeWithdrawalRejected, "ADJUSTMENT": entity.TxTypeAdjustment}
	result, ok := values[value]
	if !ok {
		return nil, fmt.Errorf("invalid wallet transaction type")
	}
	return &result, nil
}

func ParseDirection(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" || value == "CREDIT" || value == "DEBIT" {
		return value, nil
	}
	return "", fmt.Errorf("direction must be CREDIT or DEBIT")
}
