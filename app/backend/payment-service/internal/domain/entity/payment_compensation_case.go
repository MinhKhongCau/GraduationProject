package entity

import (
	"payment-service/internal/domain/vo"
	paymentdomain "payment-service/internal/payment/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentCompensationCase struct {
	ID                       uuid.UUID                            `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentOrderID           uuid.UUID                            `json:"payment_order_id" gorm:"type:uuid;not null;column:payment_order_id"`
	AppointmentID            uuid.UUID                            `json:"appointment_id" gorm:"type:uuid;not null;column:appointment_id"`
	Type                     paymentdomain.CompensationType       `json:"type" gorm:"type:varchar(40);not null;column:type"`
	Status                   paymentdomain.CompensationStatus     `json:"status" gorm:"type:varchar(30);not null;column:status"`
	ReasonCode               paymentdomain.CompensationReasonCode `json:"reason_code" gorm:"type:varchar(60);not null;column:reason_code"`
	SafeReason               string                               `json:"safe_reason" gorm:"type:varchar(500);not null;column:safe_reason"`
	GatewayOrderReference    string                               `json:"-" gorm:"type:varchar(255);not null;column:gateway_order_reference"`
	GatewayTransactionNumber string                               `json:"-" gorm:"type:varchar(255);column:gateway_transaction_number"`
	GatewayResponseCode      string                               `json:"-" gorm:"type:varchar(10);column:gateway_response_code"`
	GatewayTransactionStatus string                               `json:"-" gorm:"type:varchar(10);column:gateway_transaction_status"`
	GatewayPaymentDate       string                               `json:"-" gorm:"type:varchar(14);column:gateway_payment_date"`
	AmountVND                vo.Money                             `json:"amount_vnd" gorm:"type:bigint;not null;column:amount_vnd"`
	CreatedAt                int64                                `json:"created_at" gorm:"type:bigint;not null;column:created_at"`
	UpdatedAt                int64                                `json:"updated_at" gorm:"type:bigint;not null;column:updated_at"`
	ResolvedAt               *int64                               `json:"resolved_at,omitempty" gorm:"type:bigint;column:resolved_at"`
	ResolutionNote           *string                              `json:"resolution_note,omitempty" gorm:"type:varchar(500);column:resolution_note"`
}

func (PaymentCompensationCase) TableName() string {
	return "payment_compensation_cases"
}

func (c *PaymentCompensationCase) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
