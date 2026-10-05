package payment

import (
	"context"
	paymentdomain "payment-service/internal/domain/payment"

	"github.com/google/uuid"
)

type OutboxAttemptResult struct {
	Status          paymentdomain.OutboxStatus
	AttemptedAt     int64
	NextAttemptAt   *int64
	LastError       string
	FailureCategory paymentdomain.BookingDeliveryFailureCategory
}

type OutboxRepository interface {
	GetEligibleEvents(ctx context.Context, nowMillis int64, limit int) ([]paymentdomain.OutboxEvent, error)
	RecordAttempt(ctx context.Context, eventID uuid.UUID, result OutboxAttemptResult) (bool, error)
}
