package appointmentpostgres

import (
	"context"
	"errors"

	"booking-service/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormTx struct {
	db *gorm.DB
}

func (r *pgRepository) WithinTx(ctx context.Context, fn func(tx Tx) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormTx{db: tx})
	})
}

func (tx *gormTx) LoadAppointmentForUpdate(ctx context.Context, appointmentID string) (*domain.Appointment, error) {
	var appt domain.Appointment
	if err := tx.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("appointment_id = ?", appointmentID).
		First(&appt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &appt, nil
}

func (tx *gormTx) LoadSlotForUpdate(ctx context.Context, slotID string) (*domain.ExpertSlot, error) {
	var slot domain.ExpertSlot
	if err := tx.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("slot_id = ?", slotID).
		First(&slot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errTxRecordNotFound
		}
		return nil, err
	}
	return &slot, nil
}

func (tx *gormTx) CreateAppointment(ctx context.Context, appointment *domain.Appointment) error {
	return tx.db.WithContext(ctx).Create(appointment).Error
}

func (tx *gormTx) UpdateAppointment(ctx context.Context, appointment *domain.Appointment, updates map[string]interface{}) error {
	return tx.db.WithContext(ctx).Model(appointment).Updates(updates).Error
}

func (tx *gormTx) UpdateSlot(ctx context.Context, slotID string, expectedStatus *domain.SlotStatus, updates map[string]interface{}) (int64, error) {
	query := tx.db.WithContext(ctx).Model(&domain.ExpertSlot{}).Where("slot_id = ?", slotID)
	if expectedStatus != nil {
		query = query.Where("status = ?", *expectedStatus)
	}

	result := query.Updates(updates)
	return result.RowsAffected, result.Error
}
