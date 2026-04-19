package postgres

import (
	"booking-service/internal/domain"
	"time"

	"gorm.io/gorm"
)

// GeneratorRepository chuyên xử lý các truy vấn DB phục vụ cho việc tự động sinh lịch
type GeneratorRepository struct {
	db *gorm.DB
}

func NewGeneratorRepository(db *gorm.DB) *GeneratorRepository {
	return &GeneratorRepository{db: db}
}

// 1. Lấy lịch rảnh cố định (Availability) của chuyên gia
func (r *GeneratorRepository) GetAvailabilities(expertID string) ([]domain.Availability, error) {
	var avails []domain.Availability
	// Chỉ lấy những lịch rảnh đang được bật (is_enabled = true)
	err := r.db.Where("expert_id = ? AND is_enabled = ?", expertID, true).Find(&avails).Error
	return avails, err
}

// 2. Lấy danh sách các ca làm việc mẫu (Time Templates)
func (r *GeneratorRepository) GetTimeTemplates() ([]domain.TimeTemplate, error) {
	var templates []domain.TimeTemplate
	err := r.db.Where("is_active = ?", true).Find(&templates).Error
	return templates, err
}

// 3. Lấy danh sách ngày nghỉ xin phép (Time Off) của chuyên gia từ hôm nay trở đi
func (r *GeneratorRepository) GetTimeOffs(expertID string, fromDate time.Time) ([]domain.ExpertTimeOff, error) {
	var timeOffs []domain.ExpertTimeOff
	// Quét những lịch nghỉ chưa kết thúc tính từ thời điểm hiện tại
	err := r.db.Where("expert_id = ? AND end_datetime >= ?", expertID, fromDate).Find(&timeOffs).Error
	return timeOffs, err
}

// 4. LƯU HÀNG LOẠT (BULK INSERT): Đây là hàm quan trọng nhất để chống nghẽn CSDL
func (r *GeneratorRepository) BulkInsertSlots(slots []domain.ExpertSlot) error {
	if len(slots) == 0 {
		return nil // Nếu không có slot nào được tạo ra thì thôi, không làm gì cả
	}

	// CreateInBatches chia nhỏ mảng slots ra. Thay vì đẩy 1 phát 1000 dòng có thể gây đứt kết nối,
	// nó sẽ đẩy từng nhóm 100 dòng một. Đây là Best Practice khi làm việc với SQL!
	return r.db.CreateInBatches(slots, 100).Error
}
