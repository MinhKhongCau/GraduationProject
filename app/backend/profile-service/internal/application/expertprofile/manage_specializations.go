// File: internal/application/expertprofile/manage_specializations.go
package expertprofile

import (
	"context"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/specialization"

	"github.com/google/uuid"
)

// ManageExpertSpecializations là use case chuyên gia tự thêm/bớt từng chuyên khoa của mình.
type ManageExpertSpecializations struct {
	experts         expert.Repository
	specializations specialization.Repository
	events          EventPublisher
}

func NewManageExpertSpecializations(experts expert.Repository, specs specialization.Repository, events EventPublisher) *ManageExpertSpecializations {
	return &ManageExpertSpecializations{experts: experts, specializations: specs, events: events}
}

// Add đăng ký thêm chuyên khoa specID (idempotent nếu đã đăng ký).
func (uc *ManageExpertSpecializations) Add(ctx context.Context, authID, specID uuid.UUID) (ExpertProfileView, error) {
	e, err := uc.experts.FindByAuthID(ctx, authID)
	if err != nil {
		return ExpertProfileView{}, err
	}

	specs, err := uc.specializations.FindByIDs(ctx, []string{specID.String()})
	if err != nil {
		return ExpertProfileView{}, err
	}
	if len(specs) == 0 {
		return ExpertProfileView{}, specialization.ErrSpecializationNotFound
	}
	if err := e.AddSpecialization(specs[0]); err != nil {
		return ExpertProfileView{}, err
	}

	return save(ctx, uc.experts, uc.events, e)
}

// Remove huỷ đăng ký chuyên khoa specID.
func (uc *ManageExpertSpecializations) Remove(ctx context.Context, authID, specID uuid.UUID) (ExpertProfileView, error) {
	e, err := uc.experts.FindByAuthID(ctx, authID)
	if err != nil {
		return ExpertProfileView{}, err
	}
	if err := e.RemoveSpecialization(specID); err != nil {
		return ExpertProfileView{}, err
	}

	return save(ctx, uc.experts, uc.events, e)
}
