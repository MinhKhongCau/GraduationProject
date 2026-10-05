// File: internal/infrastructure/messaging/log_event_publisher.go
package messaging

import (
	"context"
	"log"

	"profile-service/internal/domain/expert"
)

// LogEventPublisher triển khai expertprofile.EventPublisher bằng cách ghi log.
// Chưa có consumer nào cần các event này; khi cần, thay bằng outbox/RabbitMQ publisher.
type LogEventPublisher struct{}

func (LogEventPublisher) Publish(_ context.Context, events []expert.Event) {
	for _, event := range events {
		log.Printf("profile-service: domain event %s %+v", event.EventName(), event)
	}
}
