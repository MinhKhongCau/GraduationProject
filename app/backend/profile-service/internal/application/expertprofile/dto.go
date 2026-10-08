// File: internal/application/expertprofile/dto.go
package expertprofile

import (
	"time"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/profile"
	"profile-service/internal/domain/specialization"

	"github.com/google/uuid"
)

// ExpertProfileView giữ nguyên hình dạng JSON mà các API đọc profile trả về (models.Profile kèm expert_profile).
type ExpertProfileView struct {
	ID              uuid.UUID           `json:"id"`
	Slug            string              `json:"slug"`
	UserInformation UserInformationView `json:"user_information"`
	AuthID          uuid.UUID           `json:"auth_id"`
	Role            profile.Role        `json:"role"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       *time.Time          `json:"deleted_at"`
	ExpertProfile   *ExpertDetailsView  `json:"expert_profile,omitempty"`
}

type UserInformationView struct {
	FullName    string     `json:"full_name"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string     `json:"gender"`
	PhoneNumber string     `json:"phone_number"`
	Country     string     `json:"country"`
}

type ExpertDetailsView struct {
	ProfileID            uuid.UUID            `json:"profile_id"`
	Email                string               `json:"email"`
	AvatarURL            string               `json:"avatar_url"`
	IntroductionVideoURL string               `json:"introduction_video_url"`
	Bio                  string               `json:"bio"`
	VerificationStatus   string               `json:"verification_status"`
	ManagedByAdminID     *uuid.UUID           `json:"managed_by_admin_id,omitempty"`
	VerifiedAt           *time.Time           `json:"verified_at,omitempty"`
	Specializations      []SpecializationView `json:"specializations,omitempty"`
}

type SpecializationView struct {
	SpecID      uuid.UUID `json:"spec_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Symptoms    []string  `json:"symptoms"`
	Location    string    `json:"location"`
	ImageURL    string    `json:"image_url"`
	IsActive    bool      `json:"is_active"`
}

// NewExpertProfileView map aggregate Expert sang DTO trả về cho tầng giao tiếp.
func NewExpertProfileView(e *expert.Expert) ExpertProfileView {
	s := e.Snapshot()

	specs := make([]SpecializationView, 0, len(s.Specializations))
	for _, spec := range s.Specializations {
		specs = append(specs, NewSpecializationView(spec))
	}

	return ExpertProfileView{
		ID:              s.ProfileID,
		Slug:            s.Slug,
		UserInformation: NewUserInformationView(s.UserInformation),
		AuthID:          s.AuthID,
		Role:            profile.RoleExpert,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		ExpertProfile: &ExpertDetailsView{
			ProfileID:            s.ProfileID,
			Email:                s.Details.Email,
			AvatarURL:            s.Details.AvatarURL,
			IntroductionVideoURL: s.Details.IntroductionVideoURL,
			Bio:                  s.Details.Bio,
			VerificationStatus:   s.VerificationStatus,
			ManagedByAdminID:     s.ManagerAdminID,
			VerifiedAt:           s.VerifiedAt,
			Specializations:      specs,
		},
	}
}

func NewUserInformationView(u profile.UserInformation) UserInformationView {
	return UserInformationView{
		FullName:    u.FullName,
		DateOfBirth: u.DateOfBirth,
		Gender:      string(u.Gender),
		PhoneNumber: u.PhoneNumber,
		Country:     u.Country,
	}
}

func NewSpecializationView(spec specialization.Specialization) SpecializationView {
	return SpecializationView{
		SpecID:      spec.ID,
		Code:        spec.Code,
		Name:        spec.Name,
		Slug:        spec.Slug,
		Description: spec.Description,
		Symptoms:    spec.Symptoms,
		Location:    spec.Location,
		ImageURL:    spec.ImageURL,
		IsActive:    spec.IsActive,
	}
}
