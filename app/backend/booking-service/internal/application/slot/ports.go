package slot

import (
	slotdomain "booking-service/internal/domain/slot"
	"time"
)

type Repository interface {
	GetAvailableDates(expertID string, startDate, endDate time.Time) ([]string, error)
	GetAvailableTimes(date string, expertID string) ([]SlotTimeResult, error)
	GetSlotsByExpert(expertID string, fromDate, toDate int64) ([]slotdomain.ExpertSlot, error)
	GetOverlappingSlots(expertID string, startMs, endMs int64) ([]slotdomain.ExpertSlot, error)
	BulkInsertSlots(slots []slotdomain.ExpertSlot) (int64, error)
}

// AvailableDateResult - Kết quả ngày có lịch trống trả về cho bệnh nhân
type AvailableDateResult struct {
	DateSlot string `json:"date_slot"` // Định dạng "YYYY-MM-DD"
}

// SlotTimeResult - Kết quả khung giờ trả về cho bệnh nhân (dùng Unix ms 13 số)
type SlotTimeResult struct {
	SlotID    string  `json:"slot_id"`
	Price     float64 `json:"price"`
	StartTime int64   `json:"start_time"` // Unix timestamp 13 số (ms)
	EndTime   int64   `json:"end_time"`   // Unix timestamp 13 số (ms)
}
