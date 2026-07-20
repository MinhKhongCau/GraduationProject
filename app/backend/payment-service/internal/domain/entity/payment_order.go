package entity

import (
	"payment-service/internal/domain/vo"
	paymentdomain "payment-service/internal/payment/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentOrderStatus int

const (
	OrderStatusPending PaymentOrderStatus = 1 // PENDING
	OrderStatusSuccess PaymentOrderStatus = 2 // SUCCESS
	OrderStatusFailed  PaymentOrderStatus = 3 // FAILED
	OrderStatusExpired PaymentOrderStatus = 4 // EXPIRED
)

func (s PaymentOrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "PENDING"
	case OrderStatusSuccess:
		return "SUCCESS"
	case OrderStatusFailed:
		return "FAILED"
	case OrderStatusExpired:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}

type PaymentOrder struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PayerID  uuid.UUID `json:"payer_id" gorm:"type:uuid;not null;column:payer_id"`
	ExpertID uuid.UUID `json:"expert_id" gorm:"type:uuid;not null;column:expert_id"`
	// AppointmentID là nullable — chỉ có giá trị khi order được tạo từ luồng đặt lịch.
	// Nếu Patient nạp tiền thủ công vào ví (top-up), trường này sẽ là NULL.
	AppointmentID            *uuid.UUID                         `json:"appointment_id,omitempty" gorm:"type:uuid;column:appointment_id"`
	GrossAmount              vo.Money                           `json:"gross_amount" gorm:"type:bigint;not null;column:gross_amount"`
	CommissionRate           float64                            `json:"commission_rate" gorm:"type:numeric(5,2);not null;column:commission_rate"`
	CommissionAmount         vo.Money                           `json:"commission_amount" gorm:"type:bigint;not null;column:commission_amount"`
	NetAmount                vo.Money                           `json:"net_amount" gorm:"type:bigint;not null;column:net_amount"`
	Gateway                  string                             `json:"gateway" gorm:"type:varchar(50);not null;column:gateway"` // VNPAY | MOMO | MOCK
	GatewayTxnRef            string                             `json:"gateway_txn_ref" gorm:"type:varchar(255);column:gateway_txn_ref"`
	GatewayResponseCode      string                             `json:"-" gorm:"type:varchar(10);column:gateway_response_code"`
	GatewayTransactionStatus string                             `json:"-" gorm:"type:varchar(10);column:gateway_transaction_status"`
	GatewayPaymentDate       string                             `json:"-" gorm:"type:varchar(14);column:gateway_payment_date"`
	Status                   PaymentOrderStatus                 `json:"status" gorm:"type:integer;not null;default:1;column:status"`
	GatewayCaptureStatus     paymentdomain.GatewayCaptureStatus `json:"gateway_capture_status" gorm:"type:varchar(30);not null;default:PENDING;column:gateway_capture_status"`
	FulfillmentStatus        paymentdomain.FulfillmentStatus    `json:"fulfillment_status" gorm:"type:varchar(30);not null;default:PENDING;column:fulfillment_status"`
	Released                 bool                               `json:"released" gorm:"type:boolean;not null;default:false;column:released"`
	CreatedAt                int64                              `json:"created_at" gorm:"type:bigint;not null;column:created_at"`
	ExpiresAt                int64                              `json:"expires_at" gorm:"type:bigint;column:expires_at"`
	PaidAt                   *int64                             `json:"paid_at" gorm:"type:bigint;column:paid_at"`
}

func (PaymentOrder) TableName() string {
	return "payment_orders"
}

func (p *PaymentOrder) BeforeCreate(tx *gorm.DB) error {
	if p.CreatedAt == 0 {
		p.CreatedAt = time.Now().UnixMilli()
	}
	if p.Status == 0 {
		p.Status = OrderStatusPending
	}
	if p.GatewayCaptureStatus == "" {
		p.GatewayCaptureStatus = paymentdomain.GatewayCapturePending
	}
	if p.FulfillmentStatus == "" {
		p.FulfillmentStatus = paymentdomain.FulfillmentPending
	}
	return nil
}
