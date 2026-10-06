// File: internal/infrastructure/http/schemas/user_information_schema.go
package schemas

// UserInformationRequest là thông tin cá nhân dùng chung mọi vai trò, dùng cho PUT (thay toàn bộ).
type UserInformationRequest struct {
	FullName    string `json:"full_name" binding:"required"`
	DateOfBirth string `json:"date_of_birth"` // Định dạng: YYYY-MM-DD, rỗng = không khai báo
	Gender      string `json:"gender" binding:"omitempty,oneof=MALE FEMALE OTHER"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
	Country     string `json:"country" binding:"omitempty,max=100"`
}

// PatchUserInformationRequest dùng cho PATCH: field nil = giữ nguyên, date_of_birth "" = xoá ngày sinh.
type PatchUserInformationRequest struct {
	FullName    *string `json:"full_name" binding:"omitempty,min=1"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *string `json:"gender" binding:"omitempty,oneof=MALE FEMALE OTHER"`
	PhoneNumber *string `json:"phone_number" binding:"omitempty,max=20"`
	Country     *string `json:"country" binding:"omitempty,max=100"`
}
