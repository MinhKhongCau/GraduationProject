// File: internal/schemas/patient_schema.go
package schemas

// UpsertPatientRequest dùng để tạo mới/thay thế toàn bộ (PUT) hồ sơ bệnh nhân.
type UpsertPatientRequest struct {
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email" binding:"omitempty,email"`
	AvatarURL   string `json:"avatar_url"`
	DateOfBirth string `json:"date_of_birth"` // Định dạng: YYYY-MM-DD
	Gender      string `json:"gender"`
	Address     string `json:"address"`
}

// PatchPatientRequest dùng để cập nhật một phần (PATCH) hồ sơ bệnh nhân - chỉ áp dụng các trường được gửi lên.
type PatchPatientRequest struct {
	Name        *string `json:"name"`
	PhoneNumber *string `json:"phone_number"`
	Email       *string `json:"email" binding:"omitempty,email"`
	AvatarURL   *string `json:"avatar_url"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *string `json:"gender"`
	Address     *string `json:"address"`
}
