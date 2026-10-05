// File: internal/domain/expert/verification_status.go
package expert

// VerificationStatus là trạng thái xác minh hồ sơ chuyên gia (value object).
type VerificationStatus string

const (
	StatusUnverified VerificationStatus = "UNVERIFIED"
	StatusPending    VerificationStatus = "PENDING"
	StatusVerified   VerificationStatus = "VERIFIED"
	StatusRejected   VerificationStatus = "REJECTED"
)

// ParseVerificationStatus kiểm tra giá trị đầu vào thuộc tập trạng thái hợp lệ.
func ParseVerificationStatus(value string) (VerificationStatus, error) {
	switch status := VerificationStatus(value); status {
	case StatusUnverified, StatusPending, StatusVerified, StatusRejected:
		return status, nil
	default:
		return "", ErrInvalidVerificationStatus
	}
}

// verificationStatusFromStorage khôi phục trạng thái đã lưu mà không kiểm tra lại (dữ liệu cũ được giữ
// nguyên); chỉ giá trị rỗng/chưa có hồ sơ mới rơi về mặc định UNVERIFIED.
func verificationStatusFromStorage(value string) VerificationStatus {
	if value == "" {
		return StatusUnverified
	}
	return VerificationStatus(value)
}
