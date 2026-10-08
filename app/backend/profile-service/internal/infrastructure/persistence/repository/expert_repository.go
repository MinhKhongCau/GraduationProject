// File: internal/infrastructure/persistence/repository/expert_repository.go
package repository

import (
	"context"
	"errors"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/specialization"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExpertRepository triển khai expert.Repository bằng GORM trên bảng profiles + expert_profiles.
type ExpertRepository struct {
	db *gorm.DB
}

var _ expert.Repository = (*ExpertRepository)(nil)

func NewExpertRepository(db *gorm.DB) *ExpertRepository {
	return &ExpertRepository{db: db}
}

func (r *ExpertRepository) FindByAuthID(ctx context.Context, authID uuid.UUID) (*expert.Expert, error) {
	var row models.Profile
	err := r.db.WithContext(ctx).
		Preload("ExpertProfile.Specializations").
		Where("auth_id = ? AND role = ?", authID, models.RoleExpert).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, expert.ErrExpertNotFound
	}
	if err != nil {
		return nil, err
	}
	return toExpert(&row), nil
}

func (r *ExpertRepository) Save(ctx context.Context, e *expert.Expert) error {
	s := e.Snapshot()
	row := models.ExpertProfile{
		ProfileID:            s.ProfileID,
		Email:                s.Details.Email,
		AvatarURL:            s.Details.AvatarURL,
		IntroductionVideoURL: s.Details.IntroductionVideoURL,
		Bio:                  s.Details.Bio,
		VerificationStatus:   s.VerificationStatus,
		ManagedByAdminID:     s.ManagerAdminID,
		VerifiedAt:           s.VerifiedAt,
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Profile{}).Where("id = ?", s.ProfileID).
			Updates(models.NewUserInformation(s.UserInformation).Columns()).Error; err != nil {
			return err
		}
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		if !e.SpecializationsChanged() {
			return nil
		}
		return tx.Model(&row).Association("Specializations").Replace(toSpecializationRows(s.Specializations))
	})
}

func toExpert(row *models.Profile) *expert.Expert {
	s := expert.Snapshot{
		ProfileID:       row.ID,
		AuthID:          row.AuthID,
		Slug:            row.Slug,
		UserInformation: row.UserInformation.ToDomain(),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if ep := row.ExpertProfile; ep != nil {
		s.Details = expert.Details{
			Email:                ep.Email,
			AvatarURL:            ep.AvatarURL,
			IntroductionVideoURL: ep.IntroductionVideoURL,
			Bio:                  ep.Bio,
		}
		s.VerificationStatus = ep.VerificationStatus
		s.ManagerAdminID = ep.ManagedByAdminID
		s.VerifiedAt = ep.VerifiedAt
		s.Specializations = toSpecializations(ep.Specializations)
	}
	return expert.Reconstitute(s)
}

func toSpecializations(rows []models.Specialization) []specialization.Specialization {
	if len(rows) == 0 {
		return nil
	}
	specs := make([]specialization.Specialization, 0, len(rows))
	for _, row := range rows {
		specs = append(specs, specialization.Specialization{
			ID:          row.SpecID,
			Code:        row.Code,
			Name:        row.Name,
			Slug:        row.Slug,
			Description: row.Description,
			Symptoms:    row.Symptoms,
			Location:    row.Location,
			ImageURL:    row.ImageURL,
			IsActive:    row.IsActive,
		})
	}
	return specs
}

func toSpecializationRows(specs []specialization.Specialization) []models.Specialization {
	rows := make([]models.Specialization, 0, len(specs))
	for _, spec := range specs {
		rows = append(rows, models.Specialization{
			SpecID:      spec.ID,
			Code:        spec.Code,
			Name:        spec.Name,
			Slug:        spec.Slug,
			Description: spec.Description,
			Symptoms:    spec.Symptoms,
			Location:    spec.Location,
			ImageURL:    spec.ImageURL,
			IsActive:    spec.IsActive,
		})
	}
	return rows
}
