package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxEvent struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AggregateType string    `json:"aggregate_type" gorm:"type:varchar(100);not null;column:aggregate_type"` // WALLET | WITHDRAWAL_REQUEST | PAYMENT_ORDER
	AggregateID   uuid.UUID `json:"aggregate_id" gorm:"type:uuid;not null;column:aggregate_id"`
	EventType     string    `json:"event_type" gorm:"type:varchar(100);not null;column:event_type"` // e.g. wallet.payment.received
	Payload       string    `json:"payload" gorm:"type:text;not null;column:payload"`
	Published     bool      `json:"published" gorm:"type:boolean;not null;default:false;column:published"`
	CreatedAt     int64     `json:"created_at" gorm:"type:bigint;not null;column:created_at"`
}

func (OutboxEvent) TableName() string {
	return "payment_outbox_events"
}

func (o *OutboxEvent) BeforeCreate(tx *gorm.DB) error {
	if o.CreatedAt == 0 {
		o.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}
