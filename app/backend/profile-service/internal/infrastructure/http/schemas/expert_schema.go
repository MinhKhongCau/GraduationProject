// File: internal/infrastructure/http/schemas/expert_schema.go
package schemas

// UpsertExpertRequest dùng để tạo mới/thay thế toàn bộ (PUT) hồ sơ chuyên gia.
type UpsertExpertRequest struct {
	Name                 string   `json:"name" binding:"required"`
	PhoneNumber          string   `json:"phone_number"`
	Email                string   `json:"email" binding:"omitempty,email"`
	AvatarURL            string   `json:"avatar_url"`
	IntroductionVideoURL string   `json:"introduction_video_url"`
	Bio                  string   `json:"bio"`
	SpecializationIDs    []string `json:"specialization_ids"`
}

// PatchExpertRequest dùng để cập nhật một phần (PATCH) hồ sơ chuyên gia.
// VerificationStatus chỉ được áp dụng khi người gọi là ADMIN (kiểm tra ở handler).
type PatchExpertRequest struct {
	Name                 *string  `json:"name"`
	PhoneNumber          *string  `json:"phone_number"`
	Email                *string  `json:"email" binding:"omitempty,email"`
	AvatarURL            *string  `json:"avatar_url"`
	IntroductionVideoURL *string  `json:"introduction_video_url"`
	Bio                  *string  `json:"bio"`
	VerificationStatus   *string  `json:"verification_status" binding:"omitempty,oneof=UNVERIFIED PENDING VERIFIED REJECTED"`
	SpecializationIDs    []string `json:"specialization_ids"`
}
