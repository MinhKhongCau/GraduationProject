// File: internal/infrastructure/persistence/repository/specialization_repository.go
package repository

import (
	"context"

	"profile-service/internal/domain/specialization"
	"profile-service/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// SpecializationRepository triển khai specialization.Repository bằng GORM.
type SpecializationRepository struct {
	db *gorm.DB
}

var _ specialization.Repository = (*SpecializationRepository)(nil)

func NewSpecializationRepository(db *gorm.DB) *SpecializationRepository {
	return &SpecializationRepository{db: db}
}

func (r *SpecializationRepository) FindByIDs(ctx context.Context, ids []string) ([]specialization.Specialization, error) {
	var rows []models.Specialization
	if err := r.db.WithContext(ctx).Where("spec_id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	return toSpecializations(rows), nil
}
