package domain

import "time"

// SlotStatus - Trạng thái slot khám (Enum)
type SlotStatus string

const (
	SlotStatusAvailable SlotStatus = "AVAILABLE"
	SlotStatusLocked    SlotStatus = "LOCKED"
	SlotStatusOccupied  SlotStatus = "OCCUPIED"
)

// ExpertSlot - Slot thực tế được sinh ra cho từng ngày
type ExpertSlot struct {
	SlotID          string     `json:"slot_id" gorm:"column:slot_id;primaryKey;type:uuid"`
	ExpertID        string     `json:"expert_id" gorm:"column:expert_id;type:uuid;not null;uniqueIndex:idx_expert_start_time"`
	DateSlot        time.Time  `json:"date_slot" gorm:"column:date_slot;type:date"` // Giữ time.Time nhưng lưu kiểu DATE trong Postgres
	StartTime       int64      `json:"start_time" gorm:"column:start_time;not null;uniqueIndex:idx_expert_start_time"` // Unix timestamp 13 số (ms)
	EndTime         int64      `json:"end_time" gorm:"column:end_time;not null"`                                        // Unix timestamp 13 số (ms)
	Status          SlotStatus `json:"status" gorm:"column:status;type:varchar(10);default:'AVAILABLE'"`
	Price           float64    `json:"price" gorm:"column:price;type:decimal(12,2)"`
	IsLocked        bool       `json:"is_locked" gorm:"column:is_locked"`
	LockedExpiresAt *int64     `json:"locked_expires_at" gorm:"column:locked_expires_at"` // Unix timestamp 13 số (ms), nullable
	LockedBy        *string    `json:"locked_by" gorm:"column:locked_by;type:uuid"`       // PatientID đang giữ chỗ, nullable
}

func (ExpertSlot) TableName() string { return "Booking_Expert_Slots" }
