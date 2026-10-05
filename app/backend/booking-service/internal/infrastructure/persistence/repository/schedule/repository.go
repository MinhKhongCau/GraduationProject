package repository

import (
	appschedule "booking-service/internal/application/schedule"
	appslot "booking-service/internal/application/slot"
	scheduledomain "booking-service/internal/domain/schedule"
	"booking-service/internal/domain/shared"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) appschedule.Repository {
	return &pgRepository{db: db}
}

// 1. Lấy lịch rảnh cố định (Availability) của chuyên gia
func (r *pgRepository) GetAvailabilities(expertID string) ([]scheduledomain.Availability, error) {
	var avails []scheduledomain.Availability
	// Chỉ lấy những lịch rảnh đang được bật (is_enabled = true)
	err := r.db.Where("expert_id = ? AND is_enabled = ?", expertID, true).Find(&avails).Error
	return avails, err
}

func (r *pgRepository) ListAvailabilities(filter appschedule.AvailabilityListQuery) ([]scheduledomain.Availability, int64, error) {
	query := r.db.Model(&scheduledomain.Availability{}).Where("expert_id = ?", filter.ExpertID)
	if filter.Active != nil {
		query = query.Where("is_enabled = ?", *filter.Active)
	}
	if filter.EffectiveFromMs > 0 {
		query = query.Where("effective_until IS NULL OR effective_until >= ?", filter.EffectiveFromMs)
	}
	if filter.EffectiveToMs > 0 {
		query = query.Where("effective_from < ?", filter.EffectiveToMs)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []scheduledomain.Availability
	err := query.Order("effective_from ASC, availability_id ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&items).Error
	return items, total, err
}

func (r *pgRepository) ListTimeTemplates(filter appschedule.TemplateListQuery) ([]scheduledomain.TimeTemplate, int64, error) {
	query := r.db.Model(&scheduledomain.TimeTemplate{})
	if filter.Active != nil {
		query = query.Where("is_active = ?", *filter.Active)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []scheduledomain.TimeTemplate
	err := query.Order("start_time ASC, template_id ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&items).Error
	return items, total, err
}

// 2. Lấy danh sách các ca làm việc mẫu (Time Templates) đang hoạt động
func (r *pgRepository) GetTimeTemplates() ([]scheduledomain.TimeTemplate, error) {
	var templates []scheduledomain.TimeTemplate
	err := r.db.Where("is_active = ?", true).Find(&templates).Error
	return templates, err
}

func (r *pgRepository) GetAllTimeTemplates() ([]scheduledomain.TimeTemplate, error) {
	var templates []scheduledomain.TimeTemplate
	err := r.db.Find(&templates).Error
	return templates, err
}

func (r *pgRepository) GetTimeTemplateByID(templateID string) (*scheduledomain.TimeTemplate, error) {
	var template scheduledomain.TimeTemplate
	if err := r.db.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appschedule.ErrNotFound
		}
		return nil, err
	}
	return &template, nil
}

func (r *pgRepository) GetAvailabilityByID(availID, expertID string) (*scheduledomain.Availability, error) {
	var availability scheduledomain.Availability
	if err := r.db.Where("availability_id = ? AND expert_id = ?", availID, expertID).First(&availability).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appschedule.ErrNotFound
		}
		return nil, err
	}
	return &availability, nil
}

func (r *pgRepository) GetEnabledAvailabilities(expertID string) ([]scheduledomain.Availability, error) {
	var availabilities []scheduledomain.Availability
	err := r.db.Where("expert_id = ? AND is_enabled = ?", expertID, true).Find(&availabilities).Error
	return availabilities, err
}

func (r *pgRepository) GetEnabledAvailabilitiesByTemplate(templateID string) ([]scheduledomain.Availability, error) {
	var availabilities []scheduledomain.Availability
	err := r.db.Where("template_id = ? AND is_enabled = ?", templateID, true).Find(&availabilities).Error
	return availabilities, err
}

func (r *pgRepository) GetEnabledExpertIDs() ([]string, error) {
	var expertIDs []string
	err := r.db.Model(&scheduledomain.Availability{}).Where("is_enabled = ?", true).Distinct().Pluck("expert_id", &expertIDs).Error
	return expertIDs, err
}

// CreateTimeTemplate lưu ca làm việc mẫu mới
func (r *pgRepository) CreateTimeTemplate(template *scheduledomain.TimeTemplate) error {
	return r.db.Create(template).Error
}

// CreateAvailability đăng ký cấu hình lịch rảnh mới cho chuyên gia
func (r *pgRepository) CreateAvailability(avail *scheduledomain.Availability) error {
	return r.db.Create(avail).Error
}

// UpdateAvailability cập nhật cấu hình lịch rảnh (ví dụ: đổi ca, đổi ngày, hoặc tắt)
func (r *pgRepository) UpdateAvailability(availID string, expertID string, updates map[string]interface{}) error {
	return r.db.Model(&scheduledomain.Availability{}).
		Where("availability_id = ? AND expert_id = ?", availID, expertID).
		Updates(updates).Error
}

// UpdateTemplate cập nhật ca làm việc mẫu (ví dụ: tắt - IsActive=false)
func (r *pgRepository) UpdateTemplate(templateID string, updates map[string]interface{}) error {
	return r.db.Model(&scheduledomain.TimeTemplate{}).
		Where("template_id = ?", templateID).
		Updates(updates).Error
}

func (r *pgRepository) ReconcileAvailability(availability scheduledomain.Availability, updates map[string]interface{}, candidates []slotdomain.ExpertSlot, nowMs int64) (inserted int64, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&scheduledomain.Availability{}).Where("availability_id = ? AND expert_id = ?", availability.AvailabilityID, availability.ExpertID).Updates(updates).Error; err != nil {
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

func (r *pgRepository) ReconcileTemplate(templateID string, updates map[string]interface{}, plans []appschedule.ReconciliationPlan, nowMs int64) (inserted int64, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&scheduledomain.TimeTemplate{}).Where("template_id = ?", templateID).Updates(updates).Error; err != nil {
			return err
		}
		var candidates []slotdomain.ExpertSlot
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
		expertID, availabilityID, slotdomain.SlotStatusAvailable, nowMs,
	).Delete(&slotdomain.ExpertSlot{}).Error
}

func insertReconciledSlots(tx *gorm.DB, candidates []slotdomain.ExpertSlot) (int64, error) {
	for _, candidate := range candidates {
		var existing []slotdomain.ExpertSlot
		if err := tx.Where("expert_id = ? AND start_time < ? AND end_time > ?", candidate.ExpertID, candidate.EndTime, candidate.StartTime).Find(&existing).Error; err != nil {
			return 0, err
		}
		for _, persisted := range existing {
			if persisted.StartTime == candidate.StartTime && persisted.EndTime == candidate.EndTime {
				continue
			}
			if shared.IntervalsOverlap(candidate.StartTime, candidate.EndTime, persisted.StartTime, persisted.EndTime) {
				return 0, appslot.ErrSlotOverlap
			}
		}
	}
	if len(candidates) == 0 {
		return 0, nil
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(candidates, 100)
	return result.RowsAffected, result.Error
}
