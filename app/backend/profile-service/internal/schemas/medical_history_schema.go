package schemas

import "time"

// DTO cho Request (Khi Thêm mới tiền sử bệnh)
type CreateMedicalHistoryRequest struct {
	ConditionName string `json:"condition_name" binding:"required"`
	Description   string `json:"description"`
	DiagnosedAt   string `json:"diagnosed_at" binding:"required"` // Định dạng: YYYY-MM-DD
	IsChronic     bool   `json:"is_chronic"`
}

// DTO cho Response (Khi trả về danh sách tiền sử bệnh)
type MedicalHistoryResponse struct {
	HistoryID     string    `json:"history_id"`
	ConditionName string    `json:"condition_name"`
	Description   string    `json:"description"`
	DiagnosedAt   time.Time `json:"diagnosed_at"`
	IsChronic     bool      `json:"is_chronic"`
	IsActive      bool      `json:"is_active"`
}
