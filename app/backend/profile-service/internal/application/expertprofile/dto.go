// File: internal/application/expertprofile/dto.go
package expertprofile

import (
	"time"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/profile"

	"github.com/google/uuid"
)

// ExpertProfileView giữ nguyên hình dạng JSON mà API cũ trả về (models.Profile kèm expert_profile).
type ExpertProfileView struct {
	ID            uuid.UUID          `json:"id"`
	Slug          string             `json:"slug"`
	Name          string             `json:"name"`
	AuthID        uuid.UUID          `json:"auth_id"`
	Role          profile.Role       `json:"role"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     *time.Time         `json:"deleted_at"`
	ExpertProfile *ExpertDetailsView `json:"expert_profile,omitempty"`
}

type ExpertDetailsView struct {
	ProfileID            uuid.UUID            `json:"profile_id"`
	PhoneNumber          string               `json:"phone_number"`
	Email                string               `json:"email"`
	AvatarURL            string               `json:"avatar_url"`
	IntroductionVideoURL string               `json:"introduction_video_url"`
	Bio                  string               `json:"bio"`
	VerificationStatus   string               `json:"verification_status"`
	Specializations      []SpecializationView `json:"specializations,omitempty"`
}

type SpecializationView struct {
	SpecID      uuid.UUID `json:"spec_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	IsActive    bool      `json:"is_active"`
}

// NewExpertProfileView map aggregate Expert sang DTO trả về cho tầng giao tiếp.
func NewExpertProfileView(e *expert.Expert) ExpertProfileView {
	s := e.Snapshot()

	specs := make([]SpecializationView, 0, len(s.Specializations))
	for _, spec := range s.Specializations {
		specs = append(specs, SpecializationView{
			SpecID:      spec.ID,
			Name:        spec.Name,
			Description: spec.Description,
			ImageURL:    spec.ImageURL,
			IsActive:    spec.IsActive,
		})
	}

	return ExpertProfileView{
		ID:        s.ProfileID,
		Slug:      s.Slug,
		Name:      s.Name,
		AuthID:    s.AuthID,
		Role:      profile.RoleExpert,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		ExpertProfile: &ExpertDetailsView{
			ProfileID:            s.ProfileID,
			PhoneNumber:          s.Details.PhoneNumber,
			Email:                s.Details.Email,
			AvatarURL:            s.Details.AvatarURL,
			IntroductionVideoURL: s.Details.IntroductionVideoURL,
			Bio:                  s.Details.Bio,
			VerificationStatus:   s.VerificationStatus,
			Specializations:      specs,
		},
	}
}
