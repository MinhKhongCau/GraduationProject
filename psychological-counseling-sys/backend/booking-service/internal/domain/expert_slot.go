package domain

import "time"

// ExpertSlot - Slot thực tế (Materialized View)
type ExpertSlot struct {
	SlotID          string     `json:"slot_id" gorm:"column:slot_id;primaryKey;type:uuid"`
	ExpertID        string     `json:"expert_id" gorm:"column:expert_id;type:uuid;not null"`
	DateSlot        time.Time  `json:"date_slot" gorm:"column:date_slot;type:date"`
	StartTime       time.Time  `json:"start_time" gorm:"column:start_time;type:time"`
	EndTime         time.Time  `json:"end_time" gorm:"column:end_time;type:time"`
	Status          string     `json:"status" gorm:"column:status;default:'AVAILABLE'"`
	Price           float64    `json:"price" gorm:"column:price;type:decimal(12,2)"`
	IsLocked        bool       `json:"is_locked" gorm:"column:is_locked"`
	LockedExpiresAt *time.Time `json:"locked_expires_at" gorm:"column:locked_expires_at"`
}

func (ExpertSlot) TableName() string { return "Booking_Expert_Slots" }
