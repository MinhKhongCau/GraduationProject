package outbox

import (
	"context"
	"errors"
	"payment-service/internal/domain/entity"
	paymentdomain "payment-service/internal/payment/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AttemptResult struct {
	Status        paymentdomain.OutboxStatus
	AttemptedAt   int64
	NextAttemptAt *int64
	LastError     string
}

type Repository interface {
	GetEligibleEvents(ctx context.Context, nowMillis int64, limit int) ([]entity.OutboxEvent, error)
	RecordAttempt(ctx context.Context, eventID uuid.UUID, result AttemptResult) (bool, error)
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) GetEligibleEvents(ctx context.Context, nowMillis int64, limit int) ([]entity.OutboxEvent, error) {
	var events []entity.OutboxEvent
	err := eligibleEventsQuery(r.db.WithContext(ctx), nowMillis, limit).Find(&events).Error
	return events, err
}

func eligibleEventsQuery(db *gorm.DB, nowMillis int64, limit int) *gorm.DB {
	return db.
		Where("status = ? OR (status = ? AND next_attempt_at <= ?)",
			paymentdomain.OutboxStatusPending,
			paymentdomain.OutboxStatusRetryWait,
			nowMillis,
		).
		Order("created_at ASC").
		Order("id ASC").
		Limit(limit)
}

func (r *pgRepository) RecordAttempt(ctx context.Context, eventID uuid.UUID, result AttemptResult) (bool, error) {
	var updated bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var event entity.OutboxEvent
		if err := outboxEventForUpdateQuery(tx, eventID).First(&event).Error; err != nil {
			return err
		}
		if event.Status == paymentdomain.OutboxStatusDelivered || event.Status == paymentdomain.OutboxStatusDead {
			return nil
		}
		if result.Status != paymentdomain.OutboxStatusDelivered &&
			result.Status != paymentdomain.OutboxStatusRetryWait &&
			result.Status != paymentdomain.OutboxStatusDead {
			return errors.New("invalid outbox attempt result status")
		}

		event.AttemptCount++
		event.LastAttemptAt = &result.AttemptedAt
		event.Status = result.Status
		event.NextAttemptAt = result.NextAttemptAt
		event.Published = result.Status == paymentdomain.OutboxStatusDelivered

		if result.LastError == "" {
			event.LastError = nil
		} else {
			lastError := result.LastError
			event.LastError = &lastError
		}
		if result.Status == paymentdomain.OutboxStatusDelivered {
			event.DeliveredAt = &result.AttemptedAt
		} else {
			event.DeliveredAt = nil
		}

		if err := tx.Save(&event).Error; err != nil {
			return err
		}
		updated = true
		return nil
	})
	return updated, err
}

func outboxEventForUpdateQuery(db *gorm.DB, eventID uuid.UUID) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", eventID)
}
