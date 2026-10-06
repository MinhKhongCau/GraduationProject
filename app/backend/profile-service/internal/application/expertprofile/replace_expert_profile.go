// File: internal/application/expertprofile/replace_expert_profile.go
package expertprofile

import (
	"context"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/specialization"

	"github.com/google/uuid"
)

// ReplaceExpertProfile là use case PUT hồ sơ chuyên gia.
type ReplaceExpertProfile struct {
	experts         expert.Repository
	specializations specialization.Repository
	events          EventPublisher
}

func NewReplaceExpertProfile(experts expert.Repository, specs specialization.Repository, events EventPublisher) *ReplaceExpertProfile {
	return &ReplaceExpertProfile{experts: experts, specializations: specs, events: events}
}

func (uc *ReplaceExpertProfile) Execute(ctx context.Context, cmd ReplaceCommand) (ExpertProfileView, error) {
	e, err := uc.experts.FindByAuthID(ctx, cmd.AuthID)
	if err != nil {
		return ExpertProfileView{}, err
	}

	e.ReviseUserInformation(cmd.UserInformation)
	e.ReviseDetails(expert.Details{
		Email:                cmd.Email,
		AvatarURL:            cmd.AvatarURL,
		IntroductionVideoURL: cmd.IntroductionVideoURL,
		Bio:                  cmd.Bio,
	})
	if err := assignSpecializations(ctx, uc.specializations, e, cmd.SpecializationIDs); err != nil {
		return ExpertProfileView{}, err
	}

	return save(ctx, uc.experts, uc.events, e)
}

// assignSpecializations nạp chuyên khoa theo ids và gán cho expert; ids nil = không thay đổi.
// id không tồn tại -> ErrSpecializationNotFound. Chuyên khoa ngừng hoạt động chỉ được giữ lại nếu
// chuyên gia đã đăng ký từ trước, không được đăng ký mới.
func assignSpecializations(ctx context.Context, repo specialization.Repository, e *expert.Expert, ids []string) error {
	if ids == nil {
		return nil
	}

	unique := make([]string, 0, len(ids))
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, raw := range ids {
		id, err := uuid.Parse(raw)
		if err != nil {
			return specialization.ErrSpecializationNotFound
		}
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id.String())
		}
	}

	var specs []specialization.Specialization
	if len(unique) > 0 {
		var err error
		if specs, err = repo.FindByIDs(ctx, unique); err != nil {
			return err
		}
		if len(specs) != len(unique) {
			return specialization.ErrSpecializationNotFound
		}
		for _, spec := range specs {
			if !spec.IsActive && !e.HasSpecialization(spec.ID) {
				return specialization.ErrSpecializationInactive
			}
		}
	}
	e.AssignSpecializations(specs)
	return nil
}

func save(ctx context.Context, repo expert.Repository, publisher EventPublisher, e *expert.Expert) (ExpertProfileView, error) {
	if err := repo.Save(ctx, e); err != nil {
		return ExpertProfileView{}, err
	}
	if events := e.PullEvents(); len(events) > 0 {
		publisher.Publish(ctx, events)
	}
	return NewExpertProfileView(e), nil
}
