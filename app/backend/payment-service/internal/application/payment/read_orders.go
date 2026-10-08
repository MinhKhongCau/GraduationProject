package payment

import (
	"context"
	"fmt"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"
	"strings"

	"github.com/google/uuid"
)

// PaymentOrderType phân loại đơn: thanh toán lịch hẹn hoặc nạp ví (không có appointment).
type PaymentOrderType string

const (
	PaymentOrderTypeAppointment PaymentOrderType = "APPOINTMENT"
	PaymentOrderTypeTopUp       PaymentOrderType = "TOP_UP"
)

// PaymentOrderFilter lọc đơn thanh toán. uuid.Nil / giá trị rỗng = không lọc theo trường đó.
// ScopeExperts = true giới hạn kết quả trong ExpertIDs (phạm vi Admin quản lý); ExpertIDs rỗng
// khi đó nghĩa là không có đơn nào.
type PaymentOrderFilter struct {
	PayerID           uuid.UUID
	ExpertID          uuid.UUID
	ScopeExperts      bool
	ExpertIDs         []uuid.UUID
	AppointmentID     *uuid.UUID
	Type              PaymentOrderType
	Status            *paymentdomain.PaymentOrderStatus
	FulfillmentStatus paymentdomain.FulfillmentStatus
	FromMs            int64
	ToMs              int64
	Page              readquery.PageRequest
}

// PaymentOrderTotals là số liệu tổng hợp theo bộ lọc (bỏ qua phân trang). Các tổng tiền chỉ tính
// đơn SUCCESS.
type PaymentOrderTotals struct {
	TotalOrders      int64
	PendingOrders    int64
	SuccessOrders    int64
	FailedOrders     int64
	ExpiredOrders    int64
	GrossAmount      int64
	CommissionAmount int64
	NetAmount        int64
}

type PatientOrderSummary struct {
	TotalOrders   int64 `json:"total_orders"`
	SuccessOrders int64 `json:"success_orders"`
	TotalPaid     int64 `json:"total_paid"`
	FromMs        int64 `json:"from"`
	ToMs          int64 `json:"to"`
}

type ExpertOrderSummary struct {
	TotalOrders     int64 `json:"total_orders"`
	SuccessOrders   int64 `json:"success_orders"`
	GrossTotal      int64 `json:"gross_total"`
	CommissionTotal int64 `json:"commission_total"`
	NetTotal        int64 `json:"net_total"`
	FromMs          int64 `json:"from"`
	ToMs            int64 `json:"to"`
}

type AdminOrderSummary struct {
	TotalOrders     int64 `json:"total_orders"`
	PendingOrders   int64 `json:"pending_orders"`
	SuccessOrders   int64 `json:"success_orders"`
	FailedOrders    int64 `json:"failed_orders"`
	ExpiredOrders   int64 `json:"expired_orders"`
	GrossTotal      int64 `json:"gross_total"`
	CommissionTotal int64 `json:"commission_total"`
	NetTotal        int64 `json:"net_total"`
	ManagedExperts  int   `json:"managed_experts"`
	FromMs          int64 `json:"from"`
	ToMs            int64 `json:"to"`
}

type PaymentOrderView struct {
	ID                          uuid.UUID                          `json:"id"`
	AppointmentID               *uuid.UUID                         `json:"appointment_id,omitempty"`
	PayerID                     uuid.UUID                          `json:"payer_id"`
	ExpertID                    uuid.UUID                          `json:"expert_id"`
	Expert                      *PartySummary                      `json:"expert,omitempty"`
	Appointment                 *AppointmentInfo                   `json:"appointment,omitempty"`
	Type                        PaymentOrderType                   `json:"type"`
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

// ExpertPaymentOrderView là đơn thanh toán nhìn từ phía chuyên gia (doanh thu, hoa hồng, thực nhận).
type ExpertPaymentOrderView struct {
	ID                uuid.UUID                       `json:"id"`
	AppointmentID     *uuid.UUID                      `json:"appointment_id,omitempty"`
	PayerID           uuid.UUID                       `json:"payer_id"`
	Payer             *PartySummary                   `json:"payer,omitempty"`
	Appointment       *AppointmentInfo                `json:"appointment,omitempty"`
	Type              PaymentOrderType                `json:"type"`
	GrossAmount       int64                           `json:"gross_amount"`
	CommissionRate    float64                         `json:"commission_rate"`
	CommissionAmount  int64                           `json:"commission_amount"`
	NetAmount         int64                           `json:"net_amount"`
	Status            string                          `json:"status"`
	FulfillmentStatus paymentdomain.FulfillmentStatus `json:"fulfillment_status"`
	Released          bool                            `json:"released"`
	CreatedAt         int64                           `json:"created_at"`
	PaidAt            *int64                          `json:"paid_at,omitempty"`
}

// AdminPaymentOrderView là toàn bộ thông tin đơn thanh toán cho Admin quản lý chuyên gia.
type AdminPaymentOrderView struct {
	ID                       uuid.UUID                          `json:"id"`
	AppointmentID            *uuid.UUID                         `json:"appointment_id,omitempty"`
	PayerID                  uuid.UUID                          `json:"payer_id"`
	ExpertID                 uuid.UUID                          `json:"expert_id"`
	Payer                    *PartySummary                      `json:"payer,omitempty"`
	Expert                   *PartySummary                      `json:"expert,omitempty"`
	Appointment              *AppointmentInfo                   `json:"appointment,omitempty"`
	Type                     PaymentOrderType                   `json:"type"`
	GrossAmount              int64                              `json:"gross_amount"`
	CommissionRate           float64                            `json:"commission_rate"`
	CommissionAmount         int64                              `json:"commission_amount"`
	NetAmount                int64                              `json:"net_amount"`
	Gateway                  string                             `json:"gateway"`
	GatewayTxnRef            string                             `json:"gateway_txn_ref"`
	GatewayResponseCode      string                             `json:"gateway_response_code,omitempty"`
	GatewayTransactionStatus string                             `json:"gateway_transaction_status,omitempty"`
	GatewayPaymentDate       string                             `json:"gateway_payment_date,omitempty"`
	Status                   string                             `json:"status"`
	GatewayCaptureStatus     paymentdomain.GatewayCaptureStatus `json:"gateway_capture_status"`
	FulfillmentStatus        paymentdomain.FulfillmentStatus    `json:"fulfillment_status"`
	Released                 bool                               `json:"released"`
	CreatedAt                int64                              `json:"created_at"`
	ExpiresAt                int64                              `json:"expires_at"`
	PaidAt                   *int64                             `json:"paid_at,omitempty"`
}

type PaymentOrderReader interface {
	ListPaymentOrders(ctx context.Context, filter PaymentOrderFilter) ([]paymentdomain.PaymentOrder, int64, error)
	GetPaymentOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error)
	SummarizePaymentOrders(ctx context.Context, filter PaymentOrderFilter) (PaymentOrderTotals, error)
}

// ListPaymentOrders trả về đơn thanh toán của chính bệnh nhân (filter.PayerID bắt buộc).
func (u *paymentUsecase) ListPaymentOrders(ctx context.Context, filter PaymentOrderFilter) (*readquery.Page[PaymentOrderView], error) {
	if filter.PayerID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	return listOrders(ctx, u, filter, paymentOrderView)
}

// GetPaymentOrder trả về đơn thanh toán của chính bệnh nhân.
func (u *paymentUsecase) GetPaymentOrder(ctx context.Context, payerID, orderID uuid.UUID) (*PaymentOrderView, error) {
	if payerID == uuid.Nil || orderID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	order, err := u.getOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.PayerID != payerID {
		return nil, ErrPaymentOrderForbidden
	}
	result := paymentOrderView(order, u.loadOrderContext(ctx, []paymentdomain.PaymentOrder{*order}))
	return &result, nil
}

// SummarizePatientOrders tổng hợp số đơn và tổng tiền bệnh nhân đã thanh toán trong khoảng lọc.
func (u *paymentUsecase) SummarizePatientOrders(ctx context.Context, filter PaymentOrderFilter) (*PatientOrderSummary, error) {
	if filter.PayerID == uuid.Nil {
		return nil, ErrPaymentOrderForbidden
	}
	totals, err := u.summarize(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &PatientOrderSummary{TotalOrders: totals.TotalOrders, SuccessOrders: totals.SuccessOrders, TotalPaid: totals.GrossAmount, FromMs: filter.FromMs, ToMs: filter.ToMs}, nil
}

func (u *paymentUsecase) orderReader() (PaymentOrderReader, error) {
	reader, ok := u.repo.(PaymentOrderReader)
	if !ok {
		return nil, ErrPaymentOrderReaderUnavailable
	}
	return reader, nil
}

func (u *paymentUsecase) getOrder(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	reader, err := u.orderReader()
	if err != nil {
		return nil, err
	}
	return reader.GetPaymentOrder(ctx, orderID)
}

func (u *paymentUsecase) summarize(ctx context.Context, filter PaymentOrderFilter) (PaymentOrderTotals, error) {
	if filter.ScopeExperts && len(filter.ExpertIDs) == 0 {
		return PaymentOrderTotals{}, nil
	}
	reader, err := u.orderReader()
	if err != nil {
		return PaymentOrderTotals{}, err
	}
	return reader.SummarizePaymentOrders(ctx, filter)
}

func listOrders[T any](ctx context.Context, u *paymentUsecase, filter PaymentOrderFilter, view func(*paymentdomain.PaymentOrder, orderContext) T) (*readquery.Page[T], error) {
	if filter.ScopeExperts && len(filter.ExpertIDs) == 0 {
		page := readquery.NewPage([]T{}, filter.Page, 0)
		return &page, nil
	}
	reader, err := u.orderReader()
	if err != nil {
		return nil, err
	}
	orders, total, err := reader.ListPaymentOrders(ctx, filter)
	if err != nil {
		return nil, err
	}
	oc := u.loadOrderContext(ctx, orders)
	items := make([]T, len(orders))
	for i := range orders {
		items[i] = view(&orders[i], oc)
	}
	page := readquery.NewPage(items, filter.Page, total)
	return &page, nil
}

func ParsePaymentOrderType(value string) (PaymentOrderType, error) {
	orderType := PaymentOrderType(strings.ToUpper(strings.TrimSpace(value)))
	switch orderType {
	case "", PaymentOrderTypeAppointment, PaymentOrderTypeTopUp:
		return orderType, nil
	default:
		return "", fmt.Errorf("%w: type must be APPOINTMENT or TOP_UP", ErrInvalidPaymentOrderFilter)
	}
}

func paymentOrderType(order *paymentdomain.PaymentOrder) PaymentOrderType {
	if order.AppointmentID == nil {
		return PaymentOrderTypeTopUp
	}
	return PaymentOrderTypeAppointment
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

func paymentOrderView(order *paymentdomain.PaymentOrder, oc orderContext) PaymentOrderView {
	return PaymentOrderView{ID: order.ID, AppointmentID: order.AppointmentID, PayerID: order.PayerID, ExpertID: order.ExpertID, Expert: oc.party(order.ExpertID), Appointment: oc.appointment(order.AppointmentID), Type: paymentOrderType(order), AmountVND: order.GrossAmount.Int64(), Status: order.Status.String(), FulfillmentStatus: order.FulfillmentStatus, GatewayCaptureStatus: order.GatewayCaptureStatus, Gateway: order.Gateway, GatewayTransactionReference: order.GatewayTxnRef, ExpiresAt: order.ExpiresAt, CreatedAt: order.CreatedAt, PaidAt: order.PaidAt}
}

func expertPaymentOrderView(order *paymentdomain.PaymentOrder, oc orderContext) ExpertPaymentOrderView {
	return ExpertPaymentOrderView{ID: order.ID, AppointmentID: order.AppointmentID, PayerID: order.PayerID, Payer: oc.party(order.PayerID), Appointment: oc.appointment(order.AppointmentID), Type: paymentOrderType(order), GrossAmount: order.GrossAmount.Int64(), CommissionRate: order.CommissionRate, CommissionAmount: order.CommissionAmount.Int64(), NetAmount: order.NetAmount.Int64(), Status: order.Status.String(), FulfillmentStatus: order.FulfillmentStatus, Released: order.Released, CreatedAt: order.CreatedAt, PaidAt: order.PaidAt}
}

func adminPaymentOrderView(order *paymentdomain.PaymentOrder, oc orderContext) AdminPaymentOrderView {
	return AdminPaymentOrderView{ID: order.ID, AppointmentID: order.AppointmentID, PayerID: order.PayerID, ExpertID: order.ExpertID, Payer: oc.party(order.PayerID), Expert: oc.party(order.ExpertID), Appointment: oc.appointment(order.AppointmentID), Type: paymentOrderType(order), GrossAmount: order.GrossAmount.Int64(), CommissionRate: order.CommissionRate, CommissionAmount: order.CommissionAmount.Int64(), NetAmount: order.NetAmount.Int64(), Gateway: order.Gateway, GatewayTxnRef: order.GatewayTxnRef, GatewayResponseCode: order.GatewayResponseCode, GatewayTransactionStatus: order.GatewayTransactionStatus, GatewayPaymentDate: order.GatewayPaymentDate, Status: order.Status.String(), GatewayCaptureStatus: order.GatewayCaptureStatus, FulfillmentStatus: order.FulfillmentStatus, Released: order.Released, CreatedAt: order.CreatedAt, ExpiresAt: order.ExpiresAt, PaidAt: order.PaidAt}
}
