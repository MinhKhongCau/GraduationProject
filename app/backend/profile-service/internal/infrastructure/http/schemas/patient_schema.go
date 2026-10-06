// File: internal/infrastructure/http/schemas/patient_schema.go
package schemas

// UpsertPatientRequest dùng để tạo mới/thay thế toàn bộ (PUT) hồ sơ bệnh nhân.
type UpsertPatientRequest struct {
	UserInformation UserInformationRequest `json:"user_information"`
	Email           string                 `json:"email" binding:"omitempty,email"`
	AvatarURL       string                 `json:"avatar_url"`
	Address         string                 `json:"address"`
}

// PatchPatientRequest dùng để cập nhật một phần (PATCH) hồ sơ bệnh nhân - chỉ áp dụng các trường được gửi lên.
type PatchPatientRequest struct {
	UserInformation *PatchUserInformationRequest `json:"user_information"`
	Email           *string                      `json:"email" binding:"omitempty,email"`
	AvatarURL       *string                      `json:"avatar_url"`
	Address         *string                      `json:"address"`
}
