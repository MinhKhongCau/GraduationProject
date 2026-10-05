// File: internal/application/expertprofile/ports.go
package expertprofile

import (
	"context"

	"profile-service/internal/domain/expert"
)

// EventPublisher phát domain event sau khi aggregate đã được lưu thành công.
// Triển khai hiện tại chỉ ghi log; có thể thay bằng outbox mà không đụng tới use case.
type EventPublisher interface {
	Publish(ctx context.Context, events []expert.Event)
}
