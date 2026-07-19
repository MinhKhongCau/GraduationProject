package application

import (
	"context"
	"payment-service/internal/domain/entity"
	paymentdomain "payment-service/internal/payment/domain"

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
	GetEligibleEvents(ctx context.Context, nowMillis int64, limit int) ([]entity.OutboxEvent, error)
	RecordAttempt(ctx context.Context, eventID uuid.UUID, result OutboxAttemptResult) (bool, error)
}
