package domain

import "time"

// SlotStatus - Trạng thái slot khám (Enum - lưu int trong DB)
type SlotStatus int

const (
	SlotStatusAvailable SlotStatus = 0 // "AVAILABLE"
	SlotStatusLocked    SlotStatus = 1 // "LOCKED"
	SlotStatusOccupied  SlotStatus = 2 // "OCCUPIED"
)

// String trả về label chuỗi tương ứng với giá trị enum
func (s SlotStatus) String() string {
	return [...]string{"AVAILABLE", "LOCKED", "OCCUPIED"}[s]
}

type ExpertSlot struct {
	SlotID      string     `json:"slot_id"   gorm:"column:slot_id;primaryKey;type:uuid"`
	ExpertID    string     `json:"expert_id" gorm:"column:expert_id;type:uuid;not null;uniqueIndex:idx_expert_start_time"`
	DateSlot    time.Time  `json:"date_slot" gorm:"column:date_slot;type:date"`
	StartTime   int64      `json:"start_time" gorm:"column:start_time;not null;uniqueIndex:idx_expert_start_time"` // Unix ms
	EndTime     int64      `json:"end_time"   gorm:"column:end_time;not null"`                                     // Unix ms
	Status      SlotStatus `json:"status"     gorm:"column:status;type:smallint;default:0"`
	StatusLabel string     `json:"status_label" gorm:"-"` // Tự động điền bởi AfterFind hook, không lưu DB
	Price       float64    `json:"price"      gorm:"column:price;type:decimal(12,2)"`
	IsLocked    bool       `json:"is_locked"  gorm:"column:is_locked"`

	LockedExpiresAt *int64  `json:"locked_expires_at" gorm:"column:locked_expires_at"` // Unix ms, nullable
	LockedBy        *string `json:"locked_by"         gorm:"column:locked_by;type:uuid"` // PatientID, nullable
}

func (ExpertSlot) TableName() string { return "Booking_Expert_Slots" }

// AfterFind - GORM hook: tự động chạy sau mỗi lần SELECT từ DB
// Điền StatusLabel để JSON response luôn trả về cả int lẫn string
func (s *ExpertSlot) AfterFind(tx interface{}) error {
	s.StatusLabel = s.Status.String()
	return nil
}
