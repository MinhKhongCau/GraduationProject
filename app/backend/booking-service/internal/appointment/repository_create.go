package appointment

import (
	"booking-service/internal/domain"
	"context"
	"errors"
)

// =====================================================================
// 2. TÃ¡ÂºÂ O CUÃ¡Â»ËœC HÃ¡ÂºÂ¸N MÃ¡Â»Å¡I (Transaction: KiÃ¡Â»Æ’m tra lÃ¡ÂºÂ¡i lock + Insert appointment)
// Ã„ÂÃ¡ÂºÂ£m bÃ¡ÂºÂ£o PatientID phÃ¡ÂºÂ£i khÃ¡Â»â€ºp vÃ¡Â»â€ºi ngÃ†Â°Ã¡Â»Âi Ã„â€˜ang giÃ¡Â»Â¯ chÃ¡Â»â€” (LockedBy)
// =====================================================================
func (r *pgRepository) CreateAppointment(appointment *domain.Appointment) error {
	return (&appointmentUsecase{repo: r, uow: r}).createAppointmentWithUOW(context.Background(), appointment)
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
