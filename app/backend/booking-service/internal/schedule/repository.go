package schedule

import (
	"booking-service/internal/booking/domain"
	"errors"
	"gorm.io/gorm"
)

type Repository interface {
	GetAvailabilities(expertID string) ([]domain.Availability, error)
	GetTimeTemplates() ([]domain.TimeTemplate, error)
	GetAllTimeTemplates() ([]domain.TimeTemplate, error)
	GetTimeTemplateByID(templateID string) (*domain.TimeTemplate, error)
	GetAvailabilityByID(availID, expertID string) (*domain.Availability, error)
	GetEnabledAvailabilities(expertID string) ([]domain.Availability, error)
	GetEnabledAvailabilitiesByTemplate(templateID string) ([]domain.Availability, error)
	CreateTimeTemplate(template *domain.TimeTemplate) error
	CreateAvailability(avail *domain.Availability) error
	UpdateAvailability(availID string, expertID string, updates map[string]interface{}) error
	UpdateTemplate(templateID string, updates map[string]interface{}) error
}

var ErrNotFound = errors.New("schedule record not found")

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

// 1. Lấy lịch rảnh cố định (Availability) của chuyên gia
func (r *pgRepository) GetAvailabilities(expertID string) ([]domain.Availability, error) {
	var avails []domain.Availability
	// Chỉ lấy những lịch rảnh đang được bật (is_enabled = true)
	err := r.db.Where("expert_id = ? AND is_enabled = ?", expertID, true).Find(&avails).Error
	return avails, err
}

// 2. Lấy danh sách các ca làm việc mẫu (Time Templates) đang hoạt động
func (r *pgRepository) GetTimeTemplates() ([]domain.TimeTemplate, error) {
	var templates []domain.TimeTemplate
	err := r.db.Where("is_active = ?", true).Find(&templates).Error
	return templates, err
}

func (r *pgRepository) GetAllTimeTemplates() ([]domain.TimeTemplate, error) {
	var templates []domain.TimeTemplate
	err := r.db.Find(&templates).Error
	return templates, err
}

func (r *pgRepository) GetTimeTemplateByID(templateID string) (*domain.TimeTemplate, error) {
	var template domain.TimeTemplate
	if err := r.db.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &template, nil
}

func (r *pgRepository) GetAvailabilityByID(availID, expertID string) (*domain.Availability, error) {
	var availability domain.Availability
	if err := r.db.Where("availability_id = ? AND expert_id = ?", availID, expertID).First(&availability).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &availability, nil
}

func (r *pgRepository) GetEnabledAvailabilities(expertID string) ([]domain.Availability, error) {
	var availabilities []domain.Availability
	err := r.db.Where("expert_id = ? AND is_enabled = ?", expertID, true).Find(&availabilities).Error
	return availabilities, err
}

func (r *pgRepository) GetEnabledAvailabilitiesByTemplate(templateID string) ([]domain.Availability, error) {
	var availabilities []domain.Availability
	err := r.db.Where("template_id = ? AND is_enabled = ?", templateID, true).Find(&availabilities).Error
	return availabilities, err
}

// CreateTimeTemplate lưu ca làm việc mẫu mới
func (r *pgRepository) CreateTimeTemplate(template *domain.TimeTemplate) error {
	return r.db.Create(template).Error
}

// CreateAvailability đăng ký cấu hình lịch rảnh mới cho chuyên gia
func (r *pgRepository) CreateAvailability(avail *domain.Availability) error {
	return r.db.Create(avail).Error
}

// UpdateAvailability cập nhật cấu hình lịch rảnh (ví dụ: đổi ca, đổi ngày, hoặc tắt)
func (r *pgRepository) UpdateAvailability(availID string, expertID string, updates map[string]interface{}) error {
	return r.db.Model(&domain.Availability{}).
		Where("availability_id = ? AND expert_id = ?", availID, expertID).
		Updates(updates).Error
}

// UpdateTemplate cập nhật ca làm việc mẫu (ví dụ: tắt - IsActive=false)
func (r *pgRepository) UpdateTemplate(templateID string, updates map[string]interface{}) error {
	return r.db.Model(&domain.TimeTemplate{}).
		Where("template_id = ?", templateID).
		Updates(updates).Error
}
