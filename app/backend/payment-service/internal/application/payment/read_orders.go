package payment

import (
	"context"
	"fmt"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"
	"strings"

	"github.com/google/uuid"
)

type PaymentOrderFilter struct {
	PayerID           uuid.UUID
	AppointmentID     *uuid.UUID
	Status            *paymentdomain.PaymentOrderStatus
	FulfillmentStatus paymentdomain.FulfillmentStatus
	FromMs            int64
	ToMs              int64
	Page              readquery.PageRequest
}

type PaymentOrderView struct {
	ID                          uuid.UUID                          `json:"id"`
	AppointmentID               *uuid.UUID                         `json:"appointment_id,omitempty"`
	PayerID                     uuid.UUID                          `json:"payer_id"`
	ExpertID                    uuid.UUID                          `json:"expert_id"`
	AmountVND                   int64                              `json:"amount_vnd"`
	Status                      string                             `json:"status"`
	FulfillmentStatus           paymentdomain.FulfillmentStatus    `json:"fulfillment_status"`
	GatewayCaptureStatus        paymentdomain.GatewayCaptureStatus `json:"gateway_capture_status"`
	Gateway                     string                             `json:"gateway"`
	GatewayTransactionReference string                             `json:"gateway_transaction_reference"`
	ExpiresAt                   int64                              `json:"expires_at"`
	CreatedAt                   int64                              `json:"created_at"`
	PaidAt                      *int64                             `json:"paid_at,omitempty"`
}

type PaymentOrderReader interface {
	ListPaymentOrders(ctx context.Context, filter PaymentOrderFilter) ([]paymentdomain.PaymentOrder, int64, error)
	GetPaymentOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error)
}

func (u *paymentUsecase) ListPaymentOrders(ctx context.Context, filter PaymentOrderFilter) (*readquery.Page[PaymentOrderView], error) {
	if filter.PayerID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	reader, ok := u.repo.(PaymentOrderReader)
	if !ok {
		return nil, ErrPaymentOrderReaderUnavailable
	}
	orders, total, err := reader.ListPaymentOrders(ctx, filter)
	if err != nil {
		return nil, err
	}
	items := make([]PaymentOrderView, len(orders))
	for i := range orders {
		items[i] = paymentOrderView(&orders[i])
	}
	page := readquery.NewPage(items, filter.Page, total)
	return &page, nil
}

func (u *paymentUsecase) GetPaymentOrder(ctx context.Context, payerID, orderID uuid.UUID, isAdmin bool) (*PaymentOrderView, error) {
	if payerID == uuid.Nil || orderID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	reader, ok := u.repo.(PaymentOrderReader)
	if !ok {
		return nil, ErrPaymentOrderReaderUnavailable
	}
	order, err := reader.GetPaymentOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if !isAdmin && order.PayerID != payerID {
		return nil, ErrPaymentOrderForbidden
	}
	result := paymentOrderView(order)
	return &result, nil
}

func ParsePaymentOrderStatus(value string) (*paymentdomain.PaymentOrderStatus, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}
	values := map[string]paymentdomain.PaymentOrderStatus{"PENDING": paymentdomain.OrderStatusPending, "SUCCESS": paymentdomain.OrderStatusSuccess, "FAILED": paymentdomain.OrderStatusFailed, "EXPIRED": paymentdomain.OrderStatusExpired}
	status, ok := values[value]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported payment status", ErrInvalidPaymentOrderFilter)
	}
	return &status, nil
}

func ParseFulfillmentStatus(value string) (paymentdomain.FulfillmentStatus, error) {
	status := paymentdomain.FulfillmentStatus(strings.ToUpper(strings.TrimSpace(value)))
	if status == "" {
		return "", nil
	}
	valid := map[paymentdomain.FulfillmentStatus]bool{paymentdomain.FulfillmentPending: true, paymentdomain.FulfillmentBookingConfirmed: true, paymentdomain.FulfillmentBookingFailed: true, paymentdomain.FulfillmentManualReview: true, paymentdomain.FulfillmentRefundRequired: true}
	if !valid[status] {
		return "", fmt.Errorf("%w: unsupported fulfillment status", ErrInvalidPaymentOrderFilter)
	}
	return status, nil
}

func paymentOrderView(order *paymentdomain.PaymentOrder) PaymentOrderView {
	return PaymentOrderView{ID: order.ID, AppointmentID: order.AppointmentID, PayerID: order.PayerID, ExpertID: order.ExpertID, AmountVND: order.GrossAmount.Int64(), Status: order.Status.String(), FulfillmentStatus: order.FulfillmentStatus, GatewayCaptureStatus: order.GatewayCaptureStatus, Gateway: order.Gateway, GatewayTransactionReference: order.GatewayTxnRef, ExpiresAt: order.ExpiresAt, CreatedAt: order.CreatedAt, PaidAt: order.PaidAt}
}
