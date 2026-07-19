package application

import (
	"context"
	"fmt"
	"payment-service/internal/domain/entity"
	paymentdomain "payment-service/internal/payment/domain"

	"github.com/google/uuid"
)

const (
	defaultCompensationPageSize = 20
	maximumCompensationPageSize = 100
)

type CompensationCaseFilter struct {
	Status         paymentdomain.CompensationStatus
	AppointmentID  *uuid.UUID
	PaymentOrderID *uuid.UUID
	Page           int
	Size           int
}

type CompensationCase struct {
	ID                       uuid.UUID                            `json:"id"`
	PaymentOrderID           uuid.UUID                            `json:"payment_order_id"`
	AppointmentID            uuid.UUID                            `json:"appointment_id"`
	Type                     paymentdomain.CompensationType       `json:"type"`
	Status                   paymentdomain.CompensationStatus     `json:"status"`
	PaymentStatus            string                               `json:"payment_status"`
	MoneyPaid                bool                                 `json:"money_paid"`
	GatewayCaptureStatus     paymentdomain.GatewayCaptureStatus   `json:"gateway_capture_status"`
	FulfillmentStatus        paymentdomain.FulfillmentStatus      `json:"fulfillment_status"`
	GatewayOrderReference    string                               `json:"gateway_order_reference"`
	GatewayTransactionNumber string                               `json:"gateway_transaction_number,omitempty"`
	GatewayResponseCode      string                               `json:"gateway_response_code,omitempty"`
	GatewayTransactionStatus string                               `json:"gateway_transaction_status,omitempty"`
	GatewayPaymentDate       string                               `json:"gateway_payment_date,omitempty"`
	ReasonCode               paymentdomain.CompensationReasonCode `json:"reason_code"`
	SafeReason               string                               `json:"safe_reason"`
	AmountVND                int64                                `json:"amount_vnd"`
	CreatedAt                int64                                `json:"created_at"`
	UpdatedAt                int64                                `json:"updated_at"`
	ResolvedAt               *int64                               `json:"resolved_at,omitempty"`
	ResolutionNote           *string                              `json:"resolution_note,omitempty"`
}

type CompensationCasePage struct {
	Items []CompensationCase `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

type CompensationCaseReader interface {
	ListCompensationCases(ctx context.Context, filter CompensationCaseFilter) ([]CompensationCaseRecord, int64, error)
	GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*CompensationCaseRecord, error)
}

type CompensationCaseRecord struct {
	Case                 entity.PaymentCompensationCase
	PaymentStatus        entity.PaymentOrderStatus
	GatewayCaptureStatus paymentdomain.GatewayCaptureStatus
	FulfillmentStatus    paymentdomain.FulfillmentStatus
}

func (u *paymentUsecase) ListCompensationCases(ctx context.Context, filter CompensationCaseFilter) (*CompensationCasePage, error) {
	if err := normalizeCompensationFilter(&filter); err != nil {
		return nil, err
	}
	reader, ok := u.repo.(CompensationCaseReader)
	if !ok {
		return nil, ErrCompensationReaderUnavailable
	}
	rows, total, err := reader.ListCompensationCases(ctx, filter)
	if err != nil {
		return nil, err
	}
	items := make([]CompensationCase, len(rows))
	for i := range rows {
		items[i] = compensationCaseFromRecord(&rows[i])
	}
	return &CompensationCasePage{Items: items, Total: total, Page: filter.Page, Size: filter.Size}, nil
}

func (u *paymentUsecase) GetCompensationCase(ctx context.Context, caseID uuid.UUID) (*CompensationCase, error) {
	if caseID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid case id", ErrInvalidCompensationFilter)
	}
	reader, ok := u.repo.(CompensationCaseReader)
	if !ok {
		return nil, ErrCompensationReaderUnavailable
	}
	row, err := reader.GetCompensationCase(ctx, caseID)
	if err != nil {
		return nil, err
	}
	result := compensationCaseFromRecord(row)
	return &result, nil
}

func normalizeCompensationFilter(filter *CompensationCaseFilter) error {
	if filter.Status != "" && filter.Status != paymentdomain.CompensationManualReview && filter.Status != paymentdomain.CompensationRefundRequired {
		return fmt.Errorf("%w: unsupported status", ErrInvalidCompensationFilter)
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 {
		filter.Size = defaultCompensationPageSize
	}
	if filter.Size > maximumCompensationPageSize {
		return fmt.Errorf("%w: size must not exceed %d", ErrInvalidCompensationFilter, maximumCompensationPageSize)
	}
	return nil
}

func compensationCaseFromRecord(record *CompensationCaseRecord) CompensationCase {
	row := &record.Case
	moneyPaid := record.PaymentStatus == entity.OrderStatusSuccess ||
		record.GatewayCaptureStatus == paymentdomain.GatewayCaptureSucceeded ||
		record.GatewayCaptureStatus == paymentdomain.GatewayCaptureDuplicate
	return CompensationCase{
		ID:                       row.ID,
		PaymentOrderID:           row.PaymentOrderID,
		AppointmentID:            row.AppointmentID,
		Type:                     row.Type,
		Status:                   row.Status,
		PaymentStatus:            record.PaymentStatus.String(),
		MoneyPaid:                moneyPaid,
		GatewayCaptureStatus:     record.GatewayCaptureStatus,
		FulfillmentStatus:        record.FulfillmentStatus,
		GatewayOrderReference:    row.GatewayOrderReference,
		GatewayTransactionNumber: row.GatewayTransactionNumber,
		GatewayResponseCode:      row.GatewayResponseCode,
		GatewayTransactionStatus: row.GatewayTransactionStatus,
		GatewayPaymentDate:       row.GatewayPaymentDate,
		ReasonCode:               row.ReasonCode,
		SafeReason:               row.SafeReason,
		AmountVND:                row.AmountVND.Int64(),
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		ResolvedAt:               row.ResolvedAt,
		ResolutionNote:           row.ResolutionNote,
	}
}
