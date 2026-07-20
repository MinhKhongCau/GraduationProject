package entity

import (
	paymentdomain "payment-service/internal/payment/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxEvent struct {
	ID                 uuid.UUID                                    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AggregateType      string                                       `json:"aggregate_type" gorm:"type:varchar(100);not null;column:aggregate_type"` // WALLET | WITHDRAWAL_REQUEST | PAYMENT_ORDER
	AggregateID        uuid.UUID                                    `json:"aggregate_id" gorm:"type:uuid;not null;column:aggregate_id"`
	EventType          string                                       `json:"event_type" gorm:"type:varchar(100);not null;column:event_type"` // e.g. wallet.payment.received
	Payload            string                                       `json:"payload" gorm:"type:text;not null;column:payload"`
	Published          bool                                         `json:"published" gorm:"type:boolean;not null;default:false;column:published"`
	Status             paymentdomain.OutboxStatus                   `json:"status" gorm:"type:varchar(20);not null;default:PENDING;column:status"`
	AttemptCount       int                                          `json:"attempt_count" gorm:"type:integer;not null;default:0;column:attempt_count"`
	NextAttemptAt      *int64                                       `json:"next_attempt_at,omitempty" gorm:"type:bigint;column:next_attempt_at"`
	LastAttemptAt      *int64                                       `json:"last_attempt_at,omitempty" gorm:"type:bigint;column:last_attempt_at"`
	DeliveredAt        *int64                                       `json:"delivered_at,omitempty" gorm:"type:bigint;column:delivered_at"`
	LastError          *string                                      `json:"last_error,omitempty" gorm:"type:varchar(500);column:last_error"`
	TerminalReasonCode paymentdomain.BookingDeliveryFailureCategory `json:"terminal_reason_code,omitempty" gorm:"type:varchar(50);column:terminal_reason_code"`
	CreatedAt          int64                                        `json:"created_at" gorm:"type:bigint;not null;column:created_at"`
}

func (OutboxEvent) TableName() string {
	return "payment_outbox_events"
}

func (o *OutboxEvent) BeforeCreate(tx *gorm.DB) error {
	if o.Status == "" {
		o.Status = paymentdomain.OutboxStatusPending
	}
	if o.CreatedAt == 0 {
		o.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}
