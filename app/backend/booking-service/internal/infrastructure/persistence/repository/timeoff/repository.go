package repository

import (
	apptimeoff "booking-service/internal/application/timeoff"
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	timeoffdomain "booking-service/internal/domain/timeoff"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) apptimeoff.Repository { return &pgRepository{db: db} }

func (r *pgRepository) CreateAndProcessTimeOff(item *timeoffdomain.ExpertTimeOff, force bool, nowMs int64) (conflicts []string, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := rejectDuplicateTimeOff(tx, item); err != nil {
			return err
		}
		var reconcileErr error
		conflicts, reconcileErr = reconcileTimeOff(tx, item, force, nowMs)
		if reconcileErr != nil {
			return reconcileErr
		}
		processedAt := nowMs
		item.ProcessedAt = &processedAt
		return tx.Create(item).Error
	})
	return conflicts, err
}

func (r *pgRepository) ProcessExistingTimeOff(timeOffID string, force bool, nowMs int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var item timeoffdomain.ExpertTimeOff
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("time_off_id = ?", timeOffID).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apptimeoff.ErrTimeOffNotFound
			}
			return err
		}
		if item.ProcessedAt != nil {
			return nil
		}
		if _, err := reconcileTimeOff(tx, &item, force, nowMs); err != nil {
			return err
		}
		return tx.Model(&item).Update("processed_at", nowMs).Error
	})
}

func rejectDuplicateTimeOff(tx *gorm.DB, item *timeoffdomain.ExpertTimeOff) error {
	var count int64
	err := tx.Model(&timeoffdomain.ExpertTimeOff{}).
		Where("expert_id = ? AND start_datetime < ? AND end_datetime > ?", item.ExpertID, item.EndDatetime, item.StartDatetime).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return apptimeoff.ErrDuplicate
	}
	return nil
}

func reconcileTimeOff(tx *gorm.DB, item *timeoffdomain.ExpertTimeOff, force bool, nowMs int64) ([]string, error) {
	var slots []slotdomain.ExpertSlot
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("expert_id = ? AND start_time < ? AND end_time > ?", item.ExpertID, item.EndDatetime, item.StartDatetime).
		Find(&slots).Error; err != nil {
		return nil, err
	}
	if len(slots) == 0 {
		return nil, nil
	}
	slotIDs := make([]string, 0, len(slots))
	for _, slot := range slots {
		slotIDs = append(slotIDs, slot.SlotID)
	}
	var appointments []appointmentdomain.Appointment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("slot_id IN ?", slotIDs).Find(&appointments).Error; err != nil {
		return nil, err
	}
	plan, err := timeoffdomain.PlanTimeOffReconciliation(slots, appointments, force, nowMs)
	if err != nil {
		if errors.Is(err, timeoffdomain.ErrTimeOffProtectedConflict) {
			return plan.ConflictIDs, apptimeoff.ErrConflict
		}
		return nil, err
	}

	cancelledBy := string(appointmentdomain.CancellationActorSystem)
	for _, appointmentID := range plan.CancelAppointmentIDs {
		updates := map[string]interface{}{
			"status": appointmentdomain.AppointmentStatusCancelled, "cancellation_reason": "Expert force-confirmed time-off: " + item.Reason,
			"cancelled_by": &cancelledBy, "updated_at": nowMs,
		}
		if err := tx.Model(&appointmentdomain.Appointment{}).Where("appointment_id = ? AND status = ?", appointmentID, appointmentdomain.AppointmentStatusPendingPayment).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	for _, slotID := range plan.MakeUnavailableSlotIDs {
		updates := map[string]interface{}{"status": slotdomain.SlotStatusUnavailable, "locked_by": nil, "locked_expires_at": nil, "updated_at": nowMs}
		if err := tx.Model(&slotdomain.ExpertSlot{}).Where("slot_id = ?", slotID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (r *pgRepository) GetUnprocessedTimeOffs() ([]timeoffdomain.ExpertTimeOff, error) {
	var items []timeoffdomain.ExpertTimeOff
	err := r.db.Where("processed_at IS NULL").Order("expert_id, start_datetime").Find(&items).Error
	return items, err
}

func (r *pgRepository) GetTimeOffsByExpert(expertID string) ([]timeoffdomain.ExpertTimeOff, error) {
	var items []timeoffdomain.ExpertTimeOff
	err := r.db.Where("expert_id = ?", expertID).Order("start_datetime desc").Find(&items).Error
	return items, err
}

func (r *pgRepository) DeleteTimeOff(timeOffID string, expertID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var item timeoffdomain.ExpertTimeOff
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("time_off_id = ? AND expert_id = ?", timeOffID, expertID).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apptimeoff.ErrTimeOffNotFound
			}
			return err
		}
		if err := tx.Delete(&item).Error; err != nil {
			return err
		}
		return tx.Where(
			`expert_id = ? AND status = ? AND start_time > ? AND start_time < ? AND end_time > ? AND NOT EXISTS (`+
				`SELECT 1 FROM "Booking_Expert_Time_Off" active_off WHERE active_off.expert_id = "Booking_Expert_Slots".expert_id AND active_off.start_datetime < "Booking_Expert_Slots".end_time AND active_off.end_datetime > "Booking_Expert_Slots".start_time) AND NOT EXISTS (`+
				`SELECT 1 FROM "Booking_Appointments" appointment WHERE appointment.slot_id = "Booking_Expert_Slots".slot_id)`,
			expertID, slotdomain.SlotStatusUnavailable, time.Now().UnixMilli(), item.EndDatetime, item.StartDatetime,
		).Delete(&slotdomain.ExpertSlot{}).Error
	})
}

func (r *pgRepository) GetTimeOffs(expertID string, fromDate time.Time) ([]timeoffdomain.ExpertTimeOff, error) {
	var items []timeoffdomain.ExpertTimeOff
	err := r.db.Where("expert_id = ? AND end_datetime >= ?", expertID, fromDate.UnixMilli()).Find(&items).Error
	return items, err
}

func (r *pgRepository) ListTimeOffs(filter apptimeoff.TimeOffListQuery) ([]timeoffdomain.ExpertTimeOff, int64, error) {
	query := r.db.Model(&timeoffdomain.ExpertTimeOff{}).Where("expert_id = ? AND start_datetime < ? AND end_datetime > ?", filter.ExpertID, filter.ToMs, filter.FromMs)
	if filter.Processed != nil {
		if *filter.Processed {
			query = query.Where("processed_at IS NOT NULL")
		} else {
			query = query.Where("processed_at IS NULL")
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []timeoffdomain.ExpertTimeOff
	err := query.Order("start_datetime ASC, time_off_id ASC").Limit(filter.Page.Size).Offset(filter.Page.Offset()).Find(&items).Error
	return items, total, err
}
