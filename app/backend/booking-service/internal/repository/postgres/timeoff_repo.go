package postgres

import (
	"booking-service/internal/domain"
	"time"

	"gorm.io/gorm"
)

type TimeOffRepository struct {
	db *gorm.DB
}

func NewTimeOffRepository(db *gorm.DB) *TimeOffRepository {
	return &TimeOffRepository{db: db}
}

// CreateTimeOff lưu bản ghi nghỉ phép mới
func (r *TimeOffRepository) CreateTimeOff(timeOff *domain.ExpertTimeOff) error {
	return r.db.Create(timeOff).Error
}

// GetOverlappingSlots lấy danh sách slot bị đè lên bởi khoảng thời gian nghỉ
func (r *TimeOffRepository) GetOverlappingSlots(expertID string, startMs, endMs int64) ([]domain.ExpertSlot, error) {
	var slots []domain.ExpertSlot
	err := r.db.Where("expert_id = ? AND start_time < ? AND end_time > ?", expertID, endMs, startMs).Find(&slots).Error
	return slots, err
}

// GetUnprocessedTimeOffs lấy danh sách TimeOff chưa được xử lý bởi worker
func (r *TimeOffRepository) GetUnprocessedTimeOffs() ([]domain.ExpertTimeOff, error) {
	var timeOffs []domain.ExpertTimeOff
	err := r.db.Where("processed_at IS NULL").Find(&timeOffs).Error
	return timeOffs, err
}

// MarkAsProcessed cập nhật thời gian xử lý xong
func (r *TimeOffRepository) MarkAsProcessed(timeOffID string) error {
	nowMs := time.Now().UnixMilli()
	return r.db.Model(&domain.ExpertTimeOff{}).
		Where("time_off_id = ?", timeOffID).
		Update("processed_at", nowMs).Error
}

// DeleteAvailableSlots xoá các slot đang AVAILABLE nằm trong khoảng thời gian nghỉ
func (r *TimeOffRepository) DeleteAvailableSlots(expertID string, startMs, endMs int64) error {
	return r.db.Where("expert_id = ? AND status = ? AND start_time < ? AND end_time > ?", 
		expertID, domain.SlotStatusAvailable, endMs, startMs).
		Delete(&domain.ExpertSlot{}).Error
}
