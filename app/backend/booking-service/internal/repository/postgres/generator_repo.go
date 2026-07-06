package postgres

import (
	"booking-service/internal/domain"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// 2. Lấy danh sách các ca làm việc mẫu (Time Templates) đang hoạt động
func (r *GeneratorRepository) GetTimeTemplates() ([]domain.TimeTemplate, error) {
	var templates []domain.TimeTemplate
	err := r.db.Where("is_active = ?", true).Find(&templates).Error
	return templates, err
}

// 3. Lấy danh sách ngày nghỉ của chuyên gia từ hôm nay trở đi
// So sánh với Unix timestamp 13 số (ms): end_datetime >= thời điểm bắt đầu hôm nay
func (r *GeneratorRepository) GetTimeOffs(expertID string, fromDate time.Time) ([]domain.ExpertTimeOff, error) {
	var timeOffs []domain.ExpertTimeOff
	fromMs := fromDate.UnixMilli()
	err := r.db.Where("expert_id = ? AND end_datetime >= ?", expertID, fromMs).Find(&timeOffs).Error
	return timeOffs, err
}

// 4. LƯU HÀNG LOẠT (BULK INSERT) với cơ chế chống trùng lịch (Idempotency)
// Sử dụng ON CONFLICT DO NOTHING: nếu slot (expert_id, start_time) đã tồn tại thì bỏ qua, không báo lỗi
func (r *GeneratorRepository) BulkInsertSlots(slots []domain.ExpertSlot) error {
	if len(slots) == 0 {
		return nil
	}

	// CreateInBatches + OnConflict DoNothing: chia nhỏ 100 dòng/lần và bỏ qua khi trùng
	return r.db.
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(slots, 100).Error
}

// CreateTimeTemplate lưu ca làm việc mẫu mới
func (r *GeneratorRepository) CreateTimeTemplate(template *domain.TimeTemplate) error {
	return r.db.Create(template).Error
}

// CreateAvailability đăng ký cấu hình lịch rảnh mới cho chuyên gia
func (r *GeneratorRepository) CreateAvailability(avail *domain.Availability) error {
	return r.db.Create(avail).Error
}
