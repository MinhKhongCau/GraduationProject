// File: internal/infrastructure/http/schemas/profile_schema.go
package schemas

// CreateProfileRequest là payload nội bộ do auth-service gửi sang (qua /internal/api/v1/profiles/create)
// ngay sau khi một tài khoản mới được đăng ký, để khởi tạo Profile + hồ sơ theo vai trò tương ứng.
type CreateProfileRequest struct {
	AuthID      string `json:"auth_id" binding:"required,uuid"`
	FullName    string `json:"full_name" binding:"required"`
	Role        string `json:"role" binding:"required,oneof=ADMIN EXPERT PATIENT"`
	Email       string `json:"email" binding:"omitempty,email"`
	DateOfBirth string `json:"date_of_birth"` // Định dạng: YYYY-MM-DD (tuỳ chọn)
}

// UpdateProfileRequest dùng cho PUT (thay thế toàn bộ) thông tin người dùng của Profile.
type UpdateProfileRequest struct {
	UserInformation UserInformationRequest `json:"user_information"`
}

// PatchProfileRequest dùng cho PATCH (cập nhật một phần) thông tin người dùng của Profile.
type PatchProfileRequest struct {
	UserInformation *PatchUserInformationRequest `json:"user_information"`
}

// UpsertAdminRequest dùng để tạo mới/thay thế toàn bộ (PUT) hồ sơ quản trị viên.
type UpsertAdminRequest struct {
	UserInformation UserInformationRequest `json:"user_information"`
	Email           string                 `json:"email" binding:"omitempty,email"`
	Note            string                 `json:"note"`
}

// PatchAdminRequest dùng để cập nhật một phần (PATCH) hồ sơ quản trị viên.
type PatchAdminRequest struct {
	UserInformation *PatchUserInformationRequest `json:"user_information"`
	Email           *string                      `json:"email" binding:"omitempty,email"`
	Note            *string                      `json:"note"`
}

// PaginationQuery là query param dùng chung cho các API danh sách (?page=&page_size=).
type PaginationQuery struct {
	Page     int `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}
