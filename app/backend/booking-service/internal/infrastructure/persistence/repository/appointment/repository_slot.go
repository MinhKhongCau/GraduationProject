package repository

import (
	appappointment "booking-service/internal/application/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// GetSlotByID đọc 1 slot (không khoá) cho trang xác nhận đặt lịch.
func (r *pgRepository) GetSlotByID(ctx context.Context, slotID string) (*slotdomain.ExpertSlot, error) {
	var slot slotdomain.ExpertSlot
	if err := r.db.WithContext(ctx).Where("slot_id = ?", slotID).First(&slot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appappointment.ErrBookingSlotNotHeld
		}
		return nil, err
	}
	return &slot, nil
}

// =====================================================================
// 1. KHÃ“A SLOT Táº M THá»œI (Atomic Update - Chá»‘ng Race Condition)
// Chá»‰ khÃ³a náº¿u slot Ä‘ang AVAILABLE
// Tráº£ vá» lá»—i náº¿u khÃ´ng cÃ³ dÃ²ng nÃ o Ä‘Æ°á»£c cáº­p nháº­t (slot Ä‘Ã£ bá»‹ ngÆ°á»i khÃ¡c láº¥y)
// =====================================================================
func (r *pgRepository) LockSlot(slotID string, patientID string) error {
	lock := slotdomain.NewSlotLock(patientID, time.Now().UnixMilli())

	result := withoutActiveTimeOffForSlot(r.db.Model(&slotdomain.ExpertSlot{})).
		Where("slot_id = ? AND status = ? AND start_time > ?", slotID, slotdomain.SlotStatusAvailable, time.Now().Add(5*time.Minute).UnixMilli()).
		Updates(lock.Updates())

	if result.Error != nil {
		return result.Error
	}
	// Náº¿u khÃ´ng cÃ³ dÃ²ng nÃ o bá»‹ áº£nh hÆ°á»Ÿng => slot Ä‘Ã£ bá»‹ ngÆ°á»i khÃ¡c giá»¯ hoáº·c Ä‘Ã£ Ä‘Æ°á»£c Ä‘áº·t rá»“i
	if result.RowsAffected == 0 {
		// Idempotent: bệnh nhân đang giữ chính slot này (VD quay lại trang xác nhận) thì coi như
		// khoá thành công, KHÔNG gia hạn thời gian giữ chỗ.
		var heldByPatient int64
		if err := r.db.Model(&slotdomain.ExpertSlot{}).
			Where("slot_id = ? AND status = ? AND locked_by = ? AND locked_expires_at > ?",
				slotID, slotdomain.SlotStatusLocked, patientID, time.Now().UnixMilli()).
			Count(&heldByPatient).Error; err != nil {
			return err
		}
		if heldByPatient > 0 {
			return nil
		}
		return errors.New("slot Ä‘Ã£ bá»‹ ngÆ°á»i khÃ¡c giá»¯ hoáº·c Ä‘Ã£ Ä‘Æ°á»£c Ä‘áº·t, vui lÃ²ng chá»n giá» khÃ¡c")
	}
	return nil
}
