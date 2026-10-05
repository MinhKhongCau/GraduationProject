package wallet

import (
	"context"
	"errors"
	"fmt"
	"payment-service/internal/application/readquery"
	walletdomain "payment-service/internal/domain/wallet"
	"strings"

	"github.com/google/uuid"
)

type TransactionHistoryQuery struct {
	UserID    uuid.UUID
	Type      *walletdomain.TransactionType
	Direction string
	FromMs    int64
	ToMs      int64
	Page      readquery.PageRequest
}
type transactionReader interface {
	ListTransactions(walletID uuid.UUID, query TransactionHistoryQuery) ([]walletdomain.WalletTransaction, int64, error)
}

func (u *walletUsecase) ListTransactionHistory(ctx context.Context, filter TransactionHistoryQuery) (*readquery.Page[walletdomain.WalletTransaction], error) {
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

func ParseTransactionType(value string) (*walletdomain.TransactionType, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}
	values := map[string]walletdomain.TransactionType{"PAYMENT_RECEIVED": walletdomain.TxTypePaymentReceived, "COMMISSION_DEDUCTED": walletdomain.TxTypeCommissionDeducted, "REFUND": walletdomain.TxTypeRefund, "WITHDRAWAL_LOCKED": walletdomain.TxTypeWithdrawalLocked, "WITHDRAWAL_COMPLETED": walletdomain.TxTypeWithdrawalCompleted, "WITHDRAWAL_REJECTED": walletdomain.TxTypeWithdrawalRejected, "ADJUSTMENT": walletdomain.TxTypeAdjustment}
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
