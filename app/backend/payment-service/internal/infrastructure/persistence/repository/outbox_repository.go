package repository

import (
	"context"
	"errors"
	apppayment "payment-service/internal/application/payment"
	paymentdomain "payment-service/internal/domain/payment"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type outboxRepository struct {
	db *gorm.DB
}

var errPaymentOrderNotFound = errors.New("payment order not found for outbox result")

type attemptPersistence interface {
	LoadOutboxForUpdate(ctx context.Context, eventID uuid.UUID) (*paymentdomain.OutboxEvent, error)
	SaveOutbox(ctx context.Context, event *paymentdomain.OutboxEvent) error
	LoadPaymentOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error)
	UpdatePaymentFulfillment(ctx context.Context, orderID uuid.UUID, status paymentdomain.FulfillmentStatus) error
	SaveCompensationCase(ctx context.Context, compensationCase *paymentdomain.PaymentCompensationCase) error
}

type gormAttemptPersistence struct {
	tx *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) apppayment.OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) GetEligibleEvents(ctx context.Context, nowMillis int64, limit int) ([]paymentdomain.OutboxEvent, error) {
	var events []paymentdomain.OutboxEvent
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

func (r *outboxRepository) RecordAttempt(ctx context.Context, eventID uuid.UUID, result apppayment.OutboxAttemptResult) (bool, error) {
	var updated bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		updated, err = recordAttemptWithinTx(ctx, &gormAttemptPersistence{tx: tx}, eventID, result)
		return err
	})
	return updated, err
}

func recordAttemptWithinTx(ctx context.Context, store attemptPersistence, eventID uuid.UUID, result apppayment.OutboxAttemptResult) (bool, error) {
	event, err := store.LoadOutboxForUpdate(ctx, eventID)
	if err != nil {
		return false, err
	}
	if event.Status == paymentdomain.OutboxStatusDelivered || event.Status == paymentdomain.OutboxStatusDead {
		return false, nil
	}
	if result.Status != paymentdomain.OutboxStatusDelivered &&
		result.Status != paymentdomain.OutboxStatusRetryWait &&
		result.Status != paymentdomain.OutboxStatusDead {
		return false, errors.New("invalid outbox attempt result status")
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
		event.TerminalReasonCode = ""
	} else {
		event.DeliveredAt = nil
		if result.Status == paymentdomain.OutboxStatusDead {
			event.TerminalReasonCode = result.FailureCategory
		} else {
			event.TerminalReasonCode = ""
		}
	}

	if err := store.SaveOutbox(ctx, event); err != nil {
		return false, err
	}
	if err := applyBookingDeliveryOutcome(ctx, store, event, result.AttemptedAt); err != nil {
		return false, err
	}
	return true, nil
}

func applyBookingDeliveryOutcome(ctx context.Context, store attemptPersistence, event *paymentdomain.OutboxEvent, nowMillis int64) error {
	if event.AggregateType != "PAYMENT_ORDER" ||
		(event.EventType != paymentdomain.BookingConfirmEvent && event.EventType != paymentdomain.BookingFailEvent) {
		return nil
	}

	order, err := store.LoadPaymentOrderForUpdate(ctx, event.AggregateID)
	if errors.Is(err, errPaymentOrderNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	decision := paymentdomain.PlanBookingDelivery(
		event.EventType,
		event.Status,
		order.Status == paymentdomain.OrderStatusSuccess,
		order.Status == paymentdomain.OrderStatusFailed,
		event.TerminalReasonCode,
	)
	if decision.FulfillmentStatus == "" {
		return nil
	}
	if err := store.UpdatePaymentFulfillment(ctx, order.ID, decision.FulfillmentStatus); err != nil {
		return err
	}
	if decision.Compensation == nil || order.AppointmentID == nil {
		return nil
	}

	compensationCase := paymentdomain.PaymentCompensationCase{
		ID:                       uuid.New(),
		PaymentOrderID:           order.ID,
		AppointmentID:            *order.AppointmentID,
		Type:                     decision.Compensation.Type,
		Status:                   decision.Compensation.Status,
		ReasonCode:               decision.Compensation.ReasonCode,
		SafeReason:               decision.Compensation.SafeReason,
		GatewayOrderReference:    order.ID.String(),
		GatewayTransactionNumber: order.GatewayTxnRef,
		GatewayResponseCode:      order.GatewayResponseCode,
		GatewayTransactionStatus: order.GatewayTransactionStatus,
		GatewayPaymentDate:       order.GatewayPaymentDate,
		AmountVND:                order.GrossAmount,
		CreatedAt:                nowMillis,
		UpdatedAt:                nowMillis,
	}
	return store.SaveCompensationCase(ctx, &compensationCase)
}

func (store *gormAttemptPersistence) LoadOutboxForUpdate(ctx context.Context, eventID uuid.UUID) (*paymentdomain.OutboxEvent, error) {
	var event paymentdomain.OutboxEvent
	if err := outboxEventForUpdateQuery(store.tx.WithContext(ctx), eventID).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (store *gormAttemptPersistence) SaveOutbox(ctx context.Context, event *paymentdomain.OutboxEvent) error {
	return store.tx.WithContext(ctx).Save(event).Error
}

func (store *gormAttemptPersistence) LoadPaymentOrderForUpdate(ctx context.Context, orderID uuid.UUID) (*paymentdomain.PaymentOrder, error) {
	var order paymentdomain.PaymentOrder
	err := paymentOrderForUpdateQuery(store.tx.WithContext(ctx), orderID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errPaymentOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (store *gormAttemptPersistence) UpdatePaymentFulfillment(ctx context.Context, orderID uuid.UUID, status paymentdomain.FulfillmentStatus) error {
	return paymentFulfillmentUpdateQuery(store.tx.WithContext(ctx), orderID, status).Error
}

func (store *gormAttemptPersistence) SaveCompensationCase(ctx context.Context, compensationCase *paymentdomain.PaymentCompensationCase) error {
	return compensationCaseInsertQuery(store.tx.WithContext(ctx), compensationCase).Error
}

func paymentOrderForUpdateQuery(db *gorm.DB, orderID uuid.UUID) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", orderID)
}

func paymentFulfillmentUpdateQuery(db *gorm.DB, orderID uuid.UUID, status paymentdomain.FulfillmentStatus) *gorm.DB {
	return db.Model(&paymentdomain.PaymentOrder{}).Where("id = ?", orderID).Update("fulfillment_status", status)
}

func compensationCaseInsertQuery(db *gorm.DB, compensationCase *paymentdomain.PaymentCompensationCase) *gorm.DB {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "payment_order_id"}, {Name: "reason_code"}},
		DoNothing: true,
	}).Create(compensationCase)
}

func outboxEventForUpdateQuery(db *gorm.DB, eventID uuid.UUID) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", eventID)
}
