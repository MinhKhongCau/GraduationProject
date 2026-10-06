// File: internal/infrastructure/http/schemas/expert_schema.go
package schemas

// UpsertExpertRequest dùng để tạo mới/thay thế toàn bộ (PUT) hồ sơ chuyên gia.
type UpsertExpertRequest struct {
	UserInformation      UserInformationRequest `json:"user_information"`
	Email                string                 `json:"email" binding:"omitempty,email"`
	AvatarURL            string                 `json:"avatar_url"`
	IntroductionVideoURL string                 `json:"introduction_video_url"`
	Bio                  string                 `json:"bio"`
	SpecializationIDs    []string               `json:"specialization_ids"`
}

// PatchExpertRequest dùng để cập nhật một phần (PATCH) hồ sơ chuyên gia.
// VerificationStatus chỉ được áp dụng khi người gọi là ADMIN (kiểm tra ở handler).
type PatchExpertRequest struct {
	UserInformation      *PatchUserInformationRequest `json:"user_information"`
	Email                *string                      `json:"email" binding:"omitempty,email"`
	AvatarURL            *string                      `json:"avatar_url"`
	IntroductionVideoURL *string                      `json:"introduction_video_url"`
	Bio                  *string                      `json:"bio"`
	VerificationStatus   *string                      `json:"verification_status" binding:"omitempty,oneof=UNVERIFIED PENDING VERIFIED REJECTED"`
	SpecializationIDs    []string                     `json:"specialization_ids"`
}
