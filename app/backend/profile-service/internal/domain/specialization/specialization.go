// File: internal/domain/specialization/specialization.go
package specialization

import (
	"context"

	"github.com/google/uuid"
)

// Specialization là chuyên khoa mà chuyên gia có thể đăng ký (aggregate riêng, do Admin quản lý).
type Specialization struct {
	ID          uuid.UUID
	Name        string
	Description string
	ImageURL    string
	IsActive    bool
}

// Repository là contract đọc chuyên khoa; triển khai nằm ở infrastructure/persistence.
type Repository interface {
	// FindByIDs trả về các chuyên khoa tồn tại trong ids; id không tồn tại bị bỏ qua.
	FindByIDs(ctx context.Context, ids []string) ([]Specialization, error)
}
