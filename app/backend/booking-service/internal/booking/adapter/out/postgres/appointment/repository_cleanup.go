package appointmentpostgres

import (
	"booking-service/internal/booking/domain"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =====================================================================
// 4. DON DEP CAC LOCK HET HAN (Goi boi Background Worker moi 60 giay)
// Tim cac slot bi khoa nhung da qua 15 phut -> mo khoa + huy appointment lien quan
// =====================================================================
func (r *pgRepository) CancelExpiredLocks() (int64, error) {
	nowMs := time.Now().UnixMilli()
	var totalCleaned int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. TÃ¬m táº¥t cáº£ cÃ¡c slot Ä‘Ã£ háº¿t háº¡n lock
		var expiredSlots []domain.ExpertSlot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ? AND locked_expires_at < ?",
				domain.SlotStatusLocked, nowMs).
			Find(&expiredSlots).Error; err != nil {
			return err
		}

		if len(expiredSlots) == 0 {
			return nil // KhÃ´ng cÃ³ gÃ¬ Ä‘á»ƒ dá»n
		}

		// Láº¥y danh sÃ¡ch slot_id Ä‘á»ƒ xá»­ lÃ½ hÃ ng loáº¡t
		slotIDs := make([]string, 0, len(expiredSlots))
		for _, s := range expiredSlots {
			slotIDs = append(slotIDs, s.SlotID)
		}

		// 2. Há»§y cÃ¡c Appointment PENDING_PAYMENT liÃªn quan Ä‘áº¿n slot háº¿t háº¡n
		plan := domain.PlanExpiredLockCleanup(nowMs)
		result := tx.Model(&domain.Appointment{}).
			Where("slot_id IN ? AND status = ?", slotIDs, domain.AppointmentStatusPendingPayment).
			Updates(plan.AppointmentUpdates)
		if result.Error != nil {
			return result.Error
		}

		// 3. Release only uncovered slots; covered slots remain non-bookable.
		for _, slot := range expiredSlots {
			updates, err := releasedSlotUpdatesForCoverage(tx, slot)
			if err != nil {
				return err
			}
			result = tx.Model(&domain.ExpertSlot{}).Where("slot_id = ?", slot.SlotID).Updates(updates)
			if result.Error != nil {
				return result.Error
			}
			totalCleaned += result.RowsAffected
		}
		return nil
	})

	return totalCleaned, err
}
