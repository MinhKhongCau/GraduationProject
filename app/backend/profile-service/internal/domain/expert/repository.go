// File: internal/domain/expert/repository.go
package expert

import (
	"context"

	"github.com/google/uuid"
)

// Repository là contract lưu trữ aggregate Expert; triển khai nằm ở infrastructure/persistence.
type Repository interface {
	// FindByAuthID tìm chuyên gia theo auth_id (account id bên auth-service).
	// Trả về ErrExpertNotFound nếu không có profile vai trò EXPERT.
	FindByAuthID(ctx context.Context, authID uuid.UUID) (*Expert, error)
	// Save lưu tên, hồ sơ chuyên gia và (nếu có thay đổi) danh sách chuyên khoa trong một transaction.
	Save(ctx context.Context, e *Expert) error
}
