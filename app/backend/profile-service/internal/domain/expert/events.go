// File: internal/domain/expert/events.go
package expert

import (
	"time"

	"github.com/google/uuid"
)

// Event là domain event phát sinh từ aggregate Expert.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// VerificationStatusChanged phát sinh khi Admin đổi trạng thái xác minh của chuyên gia.
type VerificationStatusChanged struct {
	ProfileID uuid.UUID
	AuthID    uuid.UUID
	From      VerificationStatus
	To        VerificationStatus
	At        time.Time
}

func (e VerificationStatusChanged) EventName() string     { return "expert.verification_status_changed" }
func (e VerificationStatusChanged) OccurredAt() time.Time { return e.At }
