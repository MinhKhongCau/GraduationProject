package timeoff

import (
	"booking-service/internal/booking/domain"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateTimeOff(timeOff *domain.ExpertTimeOff) error
	GetOverlappingSlots(expertID string, startMs, endMs int64) ([]domain.ExpertSlot, error)
	GetUnprocessedTimeOffs() ([]domain.ExpertTimeOff, error)
	MarkAsProcessed(timeOffID string) error
	DeleteAvailableSlots(expertID string, startMs, endMs int64) error
	GetTimeOffsByExpert(expertID string) ([]domain.ExpertTimeOff, error)
	DeleteTimeOff(timeOffID string, expertID string) error
	GetOverlappingAppointments(expertID string, startMs, endMs int64) ([]domain.Appointment, error)
	GetTimeOffs(expertID string, fromDate time.Time) ([]domain.ExpertTimeOff, error)
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

// CreateTimeOff lưu bản ghi nghỉ phép mới
func (r *pgRepository) CreateTimeOff(timeOff *domain.ExpertTimeOff) error {
	return r.db.Create(timeOff).Error
}

// GetOverlappingSlots lấy danh sách slot bị đè lên bởi khoảng thời gian nghỉ
func (r *pgRepository) GetOverlappingSlots(expertID string, startMs, endMs int64) ([]domain.ExpertSlot, error) {
	var slots []domain.ExpertSlot
	err := r.db.Where("expert_id = ? AND start_time < ? AND end_time > ?", expertID, endMs, startMs).Find(&slots).Error
	return slots, err
}

// GetUnprocessedTimeOffs lấy danh sách TimeOff chưa được xử lý bởi worker
func (r *pgRepository) GetUnprocessedTimeOffs() ([]domain.ExpertTimeOff, error) {
	var timeOffs []domain.ExpertTimeOff
	err := r.db.Where("processed_at IS NULL").Find(&timeOffs).Error
	return timeOffs, err
}

// MarkAsProcessed cập nhật thời gian xử lý xong
func (r *pgRepository) MarkAsProcessed(timeOffID string) error {
	nowMs := time.Now().UnixMilli()
	return r.db.Model(&domain.ExpertTimeOff{}).
		Where("time_off_id = ?", timeOffID).
		Update("processed_at", nowMs).Error
}

// DeleteAvailableSlots xoá các slot đang AVAILABLE nằm trong khoảng thời gian nghỉ
func (r *pgRepository) DeleteAvailableSlots(expertID string, startMs, endMs int64) error {
	return r.db.Where("expert_id = ? AND status = ? AND start_time < ? AND end_time > ?",
		expertID, domain.SlotStatusAvailable, endMs, startMs).
		Delete(&domain.ExpertSlot{}).Error
}

// Lấy danh sách TimeOff của chuyên gia
func (r *pgRepository) GetTimeOffsByExpert(expertID string) ([]domain.ExpertTimeOff, error) {
	var timeOffs []domain.ExpertTimeOff
	err := r.db.Where("expert_id = ?", expertID).Order("start_datetime desc").Find(&timeOffs).Error
	return timeOffs, err
}

// Xóa TimeOff (Expert huỷ đăng ký nghỉ phép)
func (r *pgRepository) DeleteTimeOff(timeOffID string, expertID string) error {
	return r.db.Where("time_off_id = ? AND expert_id = ?", timeOffID, expertID).Delete(&domain.ExpertTimeOff{}).Error
}

// Lấy danh sách các Appointment (Lịch hẹn đã đặt) bị đè lên bởi ngày nghỉ
func (r *pgRepository) GetOverlappingAppointments(expertID string, startMs, endMs int64) ([]domain.Appointment, error) {
	var appts []domain.Appointment

	// Join với bảng Slots để lấy các appointment nằm trong các slot bị đè
	err := r.db.Table("Booking_Appointments").
		Joins("JOIN \"Booking_Expert_Slots\" ON \"Booking_Appointments\".slot_id = \"Booking_Expert_Slots\".slot_id").
		Where("\"Booking_Expert_Slots\".expert_id = ? AND \"Booking_Expert_Slots\".start_time < ? AND \"Booking_Expert_Slots\".end_time > ?", expertID, endMs, startMs).
		Where("\"Booking_Appointments\".status IN ?", []domain.AppointmentStatus{domain.AppointmentStatusPendingPayment, domain.AppointmentStatusConfirmed}).
		Select("\"Booking_Appointments\".*").
		Find(&appts).Error

	return appts, err
}

// GetTimeOffs Lấy danh sách ngày nghỉ của chuyên gia từ ngày chỉ định
func (r *pgRepository) GetTimeOffs(expertID string, fromDate time.Time) ([]domain.ExpertTimeOff, error) {
	var timeOffs []domain.ExpertTimeOff
	fromMs := fromDate.UnixMilli()
	err := r.db.Where("expert_id = ? AND end_datetime >= ?", expertID, fromMs).Find(&timeOffs).Error
	return timeOffs, err
}
