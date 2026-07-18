package appointment

import (
	"booking-service/internal/domain"
	"errors"
	"time"
)

// =====================================================================
// 1. KHÃ“A SLOT Táº M THá»œI (Atomic Update - Chá»‘ng Race Condition)
// Chá»‰ khÃ³a náº¿u slot Ä‘ang AVAILABLE
// Tráº£ vá» lá»—i náº¿u khÃ´ng cÃ³ dÃ²ng nÃ o Ä‘Æ°á»£c cáº­p nháº­t (slot Ä‘Ã£ bá»‹ ngÆ°á»i khÃ¡c láº¥y)
// =====================================================================
func (r *pgRepository) LockSlot(slotID string, patientID string) error {
	lock := domain.NewSlotLock(patientID, time.Now().UnixMilli())

	result := r.db.Model(&domain.ExpertSlot{}).
		Where("slot_id = ? AND status = ?", slotID, domain.SlotStatusAvailable).
		Updates(lock.Updates())

	if result.Error != nil {
		return result.Error
	}
	// Náº¿u khÃ´ng cÃ³ dÃ²ng nÃ o bá»‹ áº£nh hÆ°á»Ÿng => slot Ä‘Ã£ bá»‹ ngÆ°á»i khÃ¡c giá»¯ hoáº·c Ä‘Ã£ Ä‘Æ°á»£c Ä‘áº·t rá»“i
	if result.RowsAffected == 0 {
		return errors.New("slot Ä‘Ã£ bá»‹ ngÆ°á»i khÃ¡c giá»¯ hoáº·c Ä‘Ã£ Ä‘Æ°á»£c Ä‘áº·t, vui lÃ²ng chá»n giá» khÃ¡c")
	}
	return nil
}
