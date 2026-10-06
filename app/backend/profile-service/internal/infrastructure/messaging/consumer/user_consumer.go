// File: internal/infrastructure/messaging/consumer/user_consumer.go
package consumer

import (
	"fmt"
	"log"
	"strings"
	"time"

	"profile-service/internal/infrastructure/http/handlers"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/messaging"
)

// HandleUserCreated reacts to auth-service's user.created event by creating
// the matching Profile + role-specific sub-record, reusing the exact same
// logic the internal REST endpoint (/internal/api/v1/profiles/create) uses.
// It is idempotent: a profile that already exists for the auth_id is treated
// as a no-op success instead of an error, so redelivery after a crash or a
// retry never fails or duplicates data.
func HandleUserCreated(event messaging.UserCreatedEvent) error {
	data := event.Data
	req := schemas.CreateProfileRequest{
		AuthID:      data.AccountID,
		FullName:    data.FullName,
		Role:        strings.ToUpper(data.Role),
		Email:       data.Email,
		DateOfBirth: normalizeDate(data.DateOfBirth),
	}

	if _, err := handlers.CreateProfileCore(req); err != nil {
		if err == handlers.ErrProfileConflict {
			log.Printf("profile-service: profile for auth_id=%s already exists, skipping (idempotent)", data.AccountID)
			return nil
		}
		return fmt.Errorf("create profile for auth_id=%s: %w", data.AccountID, err)
	}

	log.Printf("profile-service: created profile for auth_id=%s from user.created event %s", data.AccountID, event.EventID)
	return nil
}

// normalizeDate giữ phần YYYY-MM-DD của dateOfBirth từ auth-service (có thể kèm giờ) và
// bỏ qua giá trị không hợp lệ để không làm hỏng việc tạo profile.
func normalizeDate(value string) string {
	if len(value) < len("2006-01-02") {
		return ""
	}
	date := value[:len("2006-01-02")]
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return ""
	}
	return date
}
