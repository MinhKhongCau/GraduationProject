package repository

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"context"
	"errors"

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

func (tx *gormTx) LoadAppointmentForUpdate(ctx context.Context, appointmentID string) (*appointmentdomain.Appointment, error) {
	var appt appointmentdomain.Appointment
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

func (tx *gormTx) LoadSlotForUpdate(ctx context.Context, slotID string) (*slotdomain.ExpertSlot, error) {
	var slot slotdomain.ExpertSlot
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

func (tx *gormTx) IsSlotCoveredByTimeOff(ctx context.Context, slot slotdomain.ExpertSlot) (bool, error) {
	return isSlotCoveredByTimeOff(tx.db.WithContext(ctx), slot)
}

func (tx *gormTx) HasActiveAppointmentForSlot(ctx context.Context, slotID string) (bool, error) {
	var count int64
	err := tx.db.WithContext(ctx).Model(&appointmentdomain.Appointment{}).
		Where("slot_id = ? AND status IN ?", slotID, []appointmentdomain.AppointmentStatus{
			appointmentdomain.AppointmentStatusPendingPayment,
			appointmentdomain.AppointmentStatusConfirmed,
		}).
		Count(&count).Error
	return count > 0, err
}

func (tx *gormTx) CreateAppointment(ctx context.Context, appointment *appointmentdomain.Appointment) error {
	return tx.db.WithContext(ctx).Create(appointment).Error
}

func (tx *gormTx) UpdateAppointment(ctx context.Context, appointment *appointmentdomain.Appointment, updates map[string]interface{}) error {
	return tx.db.WithContext(ctx).Model(appointment).Updates(updates).Error
}

func (tx *gormTx) UpdateSlot(ctx context.Context, slotID string, expectedStatus *slotdomain.SlotStatus, updates map[string]interface{}) (int64, error) {
	query := tx.db.WithContext(ctx).Model(&slotdomain.ExpertSlot{}).Where("slot_id = ?", slotID)
	if expectedStatus != nil {
		query = query.Where("status = ?", *expectedStatus)
	}

	result := query.Updates(updates)
	return result.RowsAffected, result.Error
}
