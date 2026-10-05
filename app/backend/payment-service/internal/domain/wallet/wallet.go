package wallet

import (
	"payment-service/internal/domain/money"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	ID               uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID   `json:"user_id" gorm:"type:uuid;uniqueIndex;not null;column:user_id"`
	AvailableBalance money.Money `json:"available_balance" gorm:"type:bigint;not null;default:0;column:available_balance"`
	PendingBalance   money.Money `json:"pending_balance" gorm:"type:bigint;not null;default:0;column:pending_balance"`
	LockedBalance    money.Money `json:"locked_balance" gorm:"type:bigint;not null;default:0;column:locked_balance"`
	Version          int32       `json:"version" gorm:"type:integer;not null;default:1;column:version"`
	CreatedAt        int64       `json:"created_at" gorm:"type:bigint;not null;column:created_at"`
	UpdatedAt        int64       `json:"updated_at" gorm:"type:bigint;not null;column:updated_at"`
}

func (Wallet) TableName() string {
	return "payment_wallets"
}

func (w *Wallet) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UnixMilli()
	if w.CreatedAt == 0 {
		w.CreatedAt = now
	}
	w.UpdatedAt = now
	if w.Version == 0 {
		w.Version = 1
	}
	return nil
}

func (w *Wallet) BeforeUpdate(tx *gorm.DB) error {
	w.UpdatedAt = time.Now().UnixMilli()
	return nil
}
