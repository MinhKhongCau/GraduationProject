package postgres

import (
	"booking-service/internal/domain"
	"time"

	"gorm.io/gorm"
)

// SlotRepository xử lý các truy vấn liên quan đến việc hiển thị lịch trống cho bệnh nhân
type SlotRepository struct {
	db *gorm.DB
}

func NewSlotRepository(db *gorm.DB) *SlotRepository {
	return &SlotRepository{db: db}
}

// AvailableDateResult - Kết quả ngày có lịch trống trả về cho bệnh nhân
type AvailableDateResult struct {
	DateSlot string `json:"date_slot"` // Định dạng "YYYY-MM-DD"
}

// 1. Lấy danh sách NGÀY có lịch trống trong khoảng thời gian
// Điều kiện: status=AVAILABLE, is_locked=false, và start_time > thời điểm hiện tại (không hiển thị slot quá khứ)
func (r *SlotRepository) GetAvailableDates(startDate, endDate time.Time) ([]string, error) {
	var dates []string
	nowMs := time.Now().UnixMilli()

	// DISTINCT DATE: gom tất cả các slot trong cùng ngày thành 1 dòng
	// Lọc: chỉ lấy ngày có ít nhất 1 slot AVAILABLE, chưa bị khóa, và chưa qua (start_time > now)
	err := r.db.Table(`"Booking_Expert_Slots"`).
		Where(`date_slot >= ? AND date_slot <= ? AND status = ? AND is_locked = ? AND start_time > ?`,
			startDate, endDate, domain.SlotStatusAvailable, false, nowMs).
		Select(`DISTINCT TO_CHAR(date_slot, 'YYYY-MM-DD')`).
		Pluck(`TO_CHAR(date_slot, 'YYYY-MM-DD')`, &dates).Error

	return dates, err
}

// SlotTimeResult - Kết quả khung giờ trả về cho bệnh nhân (dùng Unix ms 13 số)
type SlotTimeResult struct {
	SlotID    string `json:"slot_id"`
	StartTime int64  `json:"start_time"` // Unix timestamp 13 số (ms)
	EndTime   int64  `json:"end_time"`   // Unix timestamp 13 số (ms)
}

// 2. Lấy danh sách KHUNG GIỜ trống của một ngày cụ thể (theo expert_id)
// Điều kiện: status=AVAILABLE, is_locked=false, start_time > now (loại slot quá khứ)
func (r *SlotRepository) GetAvailableTimes(date string, expertID string) ([]SlotTimeResult, error) {
	var times []SlotTimeResult
	nowMs := time.Now().UnixMilli()

	err := r.db.Table(`"Booking_Expert_Slots"`).
		Where(`TO_CHAR(date_slot, 'YYYY-MM-DD') = ? AND expert_id = ? AND status = ? AND is_locked = ? AND start_time > ?`,
			date, expertID, domain.SlotStatusAvailable, false, nowMs).
		Select("slot_id, start_time, end_time").
		Order("start_time ASC").
		Scan(&times).Error

	return times, err
}
