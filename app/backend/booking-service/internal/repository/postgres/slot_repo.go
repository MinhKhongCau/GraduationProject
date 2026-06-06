package postgres

import (
	"time"

	"gorm.io/gorm"
)

type SlotRepository struct {
	db *gorm.DB
}

func NewSlotRepository(db *gorm.DB) *SlotRepository {
	return &SlotRepository{db: db}
}

// 1. Lấy danh sách các NGÀY có lịch trống trong 1 khoảng thời gian
func (r *SlotRepository) GetAvailableDates(startDate, endDate time.Time) ([]string, error) {
	var dates []string

	// Dùng DISTINCT để gom các slot cùng ngày lại thành 1.
	// TO_CHAR ép kiểu Timestamp về dạng YYYY-MM-DD để dễ hiển thị lên Lịch Frontend
	err := r.db.Table("\"Booking_Expert_Slots\"").
		Where("date_slot >= ? AND date_slot <= ? AND status = ? AND is_locked = ?", startDate, endDate, "AVAILABLE", false).
		Select("DISTINCT TO_CHAR(date_slot, 'YYYY-MM-DD')").
		Pluck("TO_CHAR(date_slot, 'YYYY-MM-DD')", &dates).Error

	return dates, err
}

// Struct phụ để hứng kết quả Giờ bắt đầu - Giờ kết thúc
type TimeBlock struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// 2. Lấy danh sách các KHUNG GIỜ trống của một ngày cụ thể
func (r *SlotRepository) GetAvailableTimes(date string) ([]TimeBlock, error) {
	var times []TimeBlock

	// Tìm các slot trong ngày đó, ép giờ về dạng HH:MM
	err := r.db.Table("\"Booking_Expert_Slots\"").
		Where("TO_CHAR(date_slot, 'YYYY-MM-DD') = ? AND status = ? AND is_locked = ?", date, "AVAILABLE", false).
		Select("DISTINCT TO_CHAR(start_time, 'HH24:MI') as start_time, TO_CHAR(end_time, 'HH24:MI') as end_time").
		Order("start_time ASC").
		Scan(&times).Error

	return times, err
}
