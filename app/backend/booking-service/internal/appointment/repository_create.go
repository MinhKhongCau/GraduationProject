package appointment

import (
	"booking-service/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =====================================================================
// 2. Táº O CUá»˜C Háº¸N Má»šI (Transaction: Kiá»ƒm tra láº¡i lock + Insert appointment)
// Äáº£m báº£o PatientID pháº£i khá»›p vá»›i ngÆ°á»i Ä‘ang giá»¯ chá»— (LockedBy)
// =====================================================================
func (r *pgRepository) CreateAppointment(appointment *domain.Appointment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Láº¥y thÃ´ng tin slot vÃ  kiá»ƒm tra láº¡i tráº¡ng thÃ¡i (double-check)
		var slot domain.ExpertSlot
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("slot_id = ?", appointment.SlotID).
			First(&slot).Error; err != nil {
			return errors.New("khÃ´ng tÃ¬m tháº¥y slot: " + err.Error())
		}

		nowMs := time.Now().UnixMilli()
		if err := domain.ValidateAppointmentCreationSlot(slot, appointment, nowMs); err != nil {
			return mapAppointmentCreationSlotError(err)
		}

		// Táº¡o cuá»™c háº¹n vá»›i tráº¡ng thÃ¡i PENDING_PAYMENT
		appointment.CreatedAt = nowMs
		appointment.UpdatedAt = nowMs
		if err := tx.Create(appointment).Error; err != nil {
			return errors.New("lá»—i khi táº¡o cuá»™c háº¹n: " + err.Error())
		}

		return nil
	})
}

func mapAppointmentCreationSlotError(err error) error {
	switch {
	case errors.Is(err, domain.ErrAppointmentCreationSlotNotLockedByPatient):
		return errors.New("slot khÃ´ng Ä‘Æ°á»£c giá»¯ bá»Ÿi báº¡n, vui lÃ²ng thá»±c hiá»‡n láº¡i tá»« Ä‘áº§u")
	case errors.Is(err, domain.ErrAppointmentCreationSlotExpertMismatch):
		return errors.New("slot expert does not match appointment expert")
	case errors.Is(err, domain.ErrAppointmentCreationSlotLockExpired):
		return errors.New("phiÃªn giá»¯ chá»— Ä‘Ã£ háº¿t háº¡n 15 phÃºt, vui lÃ²ng chá»n láº¡i")
	default:
		return err
	}
}
