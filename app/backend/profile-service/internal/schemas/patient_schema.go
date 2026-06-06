// File: internal/schemas/patient_schema.go
package schemas

import "time"

// DTO dùng để nhận dữ liệu khi Bệnh nhân cập nhật profile
type UpdatePatientRequest struct {
	FullName    string `json:"full_name" binding:"required"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email" binding:"omitempty,email"`
	AvatarURL   string `json:"avatar_url"`
	DateOfBirth string `json:"date_of_birth" binding:"required"` // Định dạng: YYYY-MM-DD
	Gender      string `json:"gender"`
	Address     string `json:"address"`
}

// DTO dùng để trả dữ liệu về cho Frontend (giấu đi các trường nhạy cảm nếu có)
type PatientResponse struct {
	PatientID   string    `json:"patient_id"`
	AccountID   string    `json:"account_id"`
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Address     string    `json:"address"`
}
