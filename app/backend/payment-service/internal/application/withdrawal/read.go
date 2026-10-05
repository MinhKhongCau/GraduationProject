package withdrawal

import (
	"context"
	"errors"
	"fmt"
	"payment-service/internal/application/readquery"
	withdrawaldomain "payment-service/internal/domain/withdrawal"
	"strings"

	"github.com/google/uuid"
)

var ErrWithdrawalNotFound = errors.New("withdrawal not found")
var ErrWithdrawalForbidden = errors.New("withdrawal access denied")

type WithdrawalFilter struct {
	ActorID  uuid.UUID
	IsAdmin  bool
	ExpertID *uuid.UUID
	Status   *withdrawaldomain.WithdrawalStatus
	FromMs   int64
	ToMs     int64
	Page     readquery.PageRequest
}
type WithdrawalRecord struct {
	Request withdrawaldomain.WithdrawalRequest
	OwnerID uuid.UUID
}
type withdrawalReader interface {
	ListWithdrawals(ctx context.Context, filter WithdrawalFilter) ([]WithdrawalRecord, int64, error)
	GetWithdrawal(ctx context.Context, requestID uuid.UUID) (*WithdrawalRecord, error)
}

type WithdrawalView struct {
	ID                     uuid.UUID `json:"id"`
	WalletID               uuid.UUID `json:"wallet_id"`
	BankAccountID          uuid.UUID `json:"bank_account_id"`
	AmountVND              int64     `json:"amount_vnd"`
	Status                 string    `json:"status"`
	RequiresManualApproval bool      `json:"requires_manual_approval"`
	ApprovalNote           string    `json:"approval_note,omitempty"`
	RequestedAt            int64     `json:"requested_at"`
	ApprovedAt             *int64    `json:"approved_at,omitempty"`
	ProcessedAt            *int64    `json:"processed_at,omitempty"`
}

func (u *withdrawalUsecase) ListWithdrawals(ctx context.Context, filter WithdrawalFilter) (*readquery.Page[WithdrawalView], error) {
	if filter.ActorID == uuid.Nil {
		return nil, ErrWithdrawalForbidden
	}
	reader, ok := u.repo.(withdrawalReader)
	if !ok {
		return nil, errors.New("withdrawal reader unavailable")
	}
	rows, total, err := reader.ListWithdrawals(ctx, filter)
	if err != nil {
		return nil, err
	}
	items := make([]WithdrawalView, len(rows))
	for i := range rows {
		items[i] = withdrawalView(&rows[i].Request)
	}
	page := readquery.NewPage(items, filter.Page, total)
	return &page, nil
}

func (u *withdrawalUsecase) GetWithdrawal(ctx context.Context, actorID, requestID uuid.UUID, isAdmin bool) (*WithdrawalView, error) {
	reader, ok := u.repo.(withdrawalReader)
	if !ok {
		return nil, errors.New("withdrawal reader unavailable")
	}
	record, err := reader.GetWithdrawal(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if !isAdmin && record.OwnerID != actorID {
		return nil, ErrWithdrawalForbidden
	}
	result := withdrawalView(&record.Request)
	return &result, nil
}

func ParseWithdrawalStatus(value string) (*withdrawaldomain.WithdrawalStatus, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}
	values := map[string]withdrawaldomain.WithdrawalStatus{"PENDING": withdrawaldomain.WithdrawalStatusPending, "PENDING_APPROVAL": withdrawaldomain.WithdrawalStatusPendingApproval, "APPROVED": withdrawaldomain.WithdrawalStatusApproved, "PROCESSING": withdrawaldomain.WithdrawalStatusProcessing, "COMPLETED": withdrawaldomain.WithdrawalStatusCompleted, "REJECTED": withdrawaldomain.WithdrawalStatusRejected, "FAILED": withdrawaldomain.WithdrawalStatusFailed}
	status, ok := values[value]
	if !ok {
		return nil, fmt.Errorf("unsupported withdrawal status")
	}
	return &status, nil
}

func withdrawalView(item *withdrawaldomain.WithdrawalRequest) WithdrawalView {
	return WithdrawalView{ID: item.ID, WalletID: item.WalletID, BankAccountID: item.BankAccountID, AmountVND: item.Amount.Int64(), Status: item.Status.String(), RequiresManualApproval: item.RequiresManualApproval, ApprovalNote: item.ApprovalNote, RequestedAt: item.RequestedAt, ApprovedAt: item.ApprovedAt, ProcessedAt: item.ProcessedAt}
}
