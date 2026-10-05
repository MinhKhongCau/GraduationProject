// File: internal/application/expertprofile/replace_expert_profile.go
package expertprofile

import (
	"context"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/specialization"
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

	e.Rename(cmd.Name)
	e.ReviseDetails(expert.Details{
		PhoneNumber:          cmd.PhoneNumber,
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
func assignSpecializations(ctx context.Context, repo specialization.Repository, e *expert.Expert, ids []string) error {
	if ids == nil {
		return nil
	}
	var specs []specialization.Specialization
	if len(ids) > 0 {
		var err error
		if specs, err = repo.FindByIDs(ctx, ids); err != nil {
			return err
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
