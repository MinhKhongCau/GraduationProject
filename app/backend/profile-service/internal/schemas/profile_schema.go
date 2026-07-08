// File: internal/schemas/profile_schema.go
package schemas

import "time"

// CreateProfileRequest là payload nội bộ do auth-service gửi sang (qua /internal/api/v1/profiles/create)
// ngay sau khi một tài khoản mới được đăng ký, để khởi tạo Profile + hồ sơ theo vai trò tương ứng.
type CreateProfileRequest struct {
	AuthID string `json:"auth_id" binding:"required,uuid"`
	Name   string `json:"name" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=ADMIN EXPERT PATIENT"`
	Email  string `json:"email" binding:"omitempty,email"`
}

// UpdateProfileRequest dùng cho PUT (thay thế toàn bộ) trường chung của Profile.
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required"`
}

// PatchProfileRequest dùng cho PATCH (cập nhật một phần) trường chung của Profile.
type PatchProfileRequest struct {
	Name *string `json:"name"`
}

// ProfileResponse là hình dạng trả về cho phần thông tin gốc dùng chung mọi vai trò.
type ProfileResponse struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	AuthID    string    `json:"auth_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpsertAdminRequest dùng để tạo mới/thay thế toàn bộ (PUT) hồ sơ quản trị viên.
type UpsertAdminRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"omitempty,email"`
	Note  string `json:"note"`
}

// PatchAdminRequest dùng để cập nhật một phần (PATCH) hồ sơ quản trị viên.
type PatchAdminRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email" binding:"omitempty,email"`
	Note  *string `json:"note"`
}

// PaginationQuery là query param dùng chung cho các API danh sách (?page=&page_size=).
type PaginationQuery struct {
	Page     int `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}

// PaginatedResponse bọc danh sách kết quả kèm thông tin phân trang.
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int64       `json:"total_pages"`
}
