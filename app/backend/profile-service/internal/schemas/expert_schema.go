package schemas

// DTO khi tạo mới/cập nhật Chuyên gia
type CreateExpertRequest struct {
	FullName             string `json:"full_name" binding:"required"`
	PhoneNumber          string `json:"phone_number"`
	Email                string `json:"email" binding:"omitempty,email"`
	AvatarURL            string `json:"avatar_url"`
	IntroductionVideoURL string `json:"introduction_video_url"`
	Bio                  string `json:"bio"`
	// Mảng chứa các ID Chuyên khoa mà bác sĩ này thuộc về
	SpecializationIDs []string `json:"specialization_ids" binding:"required"`
}
