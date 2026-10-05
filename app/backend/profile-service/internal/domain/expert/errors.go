// File: internal/domain/expert/errors.go
package expert

import "errors"

var (
	// ErrExpertNotFound giữ nguyên chuỗi lỗi "record not found" mà API cũ trả về ở trường error.
	ErrExpertNotFound            = errors.New("record not found")
	ErrInvalidVerificationStatus = errors.New("invalid verification status")
	ErrVerificationRequiresAdmin = errors.New("only admin can change verification status")
)
