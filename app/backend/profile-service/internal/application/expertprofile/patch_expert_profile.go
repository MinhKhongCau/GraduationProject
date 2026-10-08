// File: internal/application/expertprofile/patch_expert_profile.go
package expertprofile

import (
	"context"
	"time"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/specialization"
)

// PatchExpertProfile là use case PATCH hồ sơ chuyên gia.
type PatchExpertProfile struct {
	experts         expert.Repository
	specializations specialization.Repository
	events          EventPublisher
}

func NewPatchExpertProfile(experts expert.Repository, specs specialization.Repository, events EventPublisher) *PatchExpertProfile {
	return &PatchExpertProfile{experts: experts, specializations: specs, events: events}
}

func (uc *PatchExpertProfile) Execute(ctx context.Context, cmd PatchCommand) (ExpertProfileView, error) {
	e, err := uc.experts.FindByAuthID(ctx, cmd.AuthID)
	if err != nil {
		return ExpertProfileView{}, err
	}

	e.ReviseUserInformation(e.UserInformation().Apply(cmd.UserInformation))

	details := e.Details()
	overwrite(&details.Email, cmd.Email)
	overwrite(&details.AvatarURL, cmd.AvatarURL)
	overwrite(&details.IntroductionVideoURL, cmd.IntroductionVideoURL)
	overwrite(&details.Bio, cmd.Bio)
	e.ReviseDetails(details)

	if cmd.VerificationStatus != nil {
		status, err := expert.ParseVerificationStatus(*cmd.VerificationStatus)
		if err != nil {
			return ExpertProfileView{}, err
		}
		if err := e.ChangeVerificationStatus(status, cmd.ActorRole, cmd.ActorID, time.Now()); err != nil {
			return ExpertProfileView{}, err
		}
	}

	if err := assignSpecializations(ctx, uc.specializations, e, cmd.SpecializationIDs); err != nil {
		return ExpertProfileView{}, err
	}

	return save(ctx, uc.experts, uc.events, e)
}

func overwrite(dst *string, value *string) {
	if value != nil {
		*dst = *value
	}
}
