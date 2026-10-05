package repository

import (
	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"context"
)

// =====================================================================
// 2. TÃ¡ÂºÂ O CUÃ¡Â»ËœC HÃ¡ÂºÂ¸N MÃ¡Â»Å¡I (Transaction: KiÃ¡Â»Æ’m tra lÃ¡ÂºÂ¡i lock + Insert appointment)
// Ã„ÂÃ¡ÂºÂ£m bÃ¡ÂºÂ£o PatientID phÃ¡ÂºÂ£i khÃ¡Â»â€ºp vÃ¡Â»â€ºi ngÃ†Â°Ã¡Â»Âi Ã„â€˜ang giÃ¡Â»Â¯ chÃ¡Â»â€” (LockedBy)
// =====================================================================
func (r *pgRepository) CreateAppointment(appointment *appointmentdomain.Appointment) error {
	return appappointment.CreateAppointmentWithUnitOfWork(context.Background(), r, r, appointment)
}
