package schedule

import (
	"booking-service/internal/booking/domain"
	"booking-service/internal/slot"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReconciliationPlan struct {
	Availability domain.Availability
	Candidates   []domain.ExpertSlot
}

type Repository interface {
	GetAvailabilities(expertID string) ([]domain.Availability, error)
	GetTimeTemplates() ([]domain.TimeTemplate, error)
	GetAllTimeTemplates() ([]domain.TimeTemplate, error)
	GetTimeTemplateByID(templateID string) (*domain.TimeTemplate, error)
	GetAvailabilityByID(availID, expertID string) (*domain.Availability, error)
	GetEnabledAvailabilities(expertID string) ([]domain.Availability, error)
	GetEnabledAvailabilitiesByTemplate(templateID string) ([]domain.Availability, error)
	GetEnabledExpertIDs() ([]string, error)
	CreateTimeTemplate(template *domain.TimeTemplate) error
	CreateAvailability(avail *domain.Availability) error
	UpdateAvailability(availID string, expertID string, updates map[string]interface{}) error
	UpdateTemplate(templateID string, updates map[string]interface{}) error
	ReconcileAvailability(availability domain.Availability, updates map[string]interface{}, candidates []domain.ExpertSlot, nowMs int64) (int64, error)
	ReconcileTemplate(templateID string, updates map[string]interface{}, plans []ReconciliationPlan, nowMs int64) (int64, error)
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

func (r *pgRepository) GetEnabledExpertIDs() ([]string, error) {
	var expertIDs []string
	err := r.db.Model(&domain.Availability{}).Where("is_enabled = ?", true).Distinct().Pluck("expert_id", &expertIDs).Error
	return expertIDs, err
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

func (r *pgRepository) ReconcileAvailability(availability domain.Availability, updates map[string]interface{}, candidates []domain.ExpertSlot, nowMs int64) (inserted int64, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Availability{}).Where("availability_id = ? AND expert_id = ?", availability.AvailabilityID, availability.ExpertID).Updates(updates).Error; err != nil {
			return err
		}
		if err := deleteReconciledSlots(tx, availability.ExpertID, availability.AvailabilityID, nowMs); err != nil {
			return err
		}
		var insertErr error
		inserted, insertErr = insertReconciledSlots(tx, candidates)
		return insertErr
	})
	return inserted, err
}

func (r *pgRepository) ReconcileTemplate(templateID string, updates map[string]interface{}, plans []ReconciliationPlan, nowMs int64) (inserted int64, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.TimeTemplate{}).Where("template_id = ?", templateID).Updates(updates).Error; err != nil {
			return err
		}
		var candidates []domain.ExpertSlot
		for _, plan := range plans {
			if err := deleteReconciledSlots(tx, plan.Availability.ExpertID, plan.Availability.AvailabilityID, nowMs); err != nil {
				return err
			}
			candidates = append(candidates, plan.Candidates...)
		}
		var insertErr error
		inserted, insertErr = insertReconciledSlots(tx, candidates)
		return insertErr
	})
	return inserted, err
}

func deleteReconciledSlots(tx *gorm.DB, expertID, availabilityID string, nowMs int64) error {
	return tx.Where(
		`expert_id = ? AND availability_id = ? AND status = ? AND start_time > ? AND NOT EXISTS (`+
			`SELECT 1 FROM "Booking_Appointments" appointment WHERE appointment.slot_id = "Booking_Expert_Slots".slot_id)`,
		expertID, availabilityID, domain.SlotStatusAvailable, nowMs,
	).Delete(&domain.ExpertSlot{}).Error
}

func insertReconciledSlots(tx *gorm.DB, candidates []domain.ExpertSlot) (int64, error) {
	for _, candidate := range candidates {
		var existing []domain.ExpertSlot
		if err := tx.Where("expert_id = ? AND start_time < ? AND end_time > ?", candidate.ExpertID, candidate.EndTime, candidate.StartTime).Find(&existing).Error; err != nil {
			return 0, err
		}
		for _, persisted := range existing {
			if persisted.StartTime == candidate.StartTime && persisted.EndTime == candidate.EndTime {
				continue
			}
			if domain.IntervalsOverlap(candidate.StartTime, candidate.EndTime, persisted.StartTime, persisted.EndTime) {
				return 0, slot.ErrSlotOverlap
			}
		}
	}
	if len(candidates) == 0 {
		return 0, nil
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(candidates, 100)
	return result.RowsAffected, result.Error
}
