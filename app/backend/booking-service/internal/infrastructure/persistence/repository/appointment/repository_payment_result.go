package repository

import appappointment "booking-service/internal/application/appointment"

// =====================================================================
// 3. XÃƒÂC NHÃ¡ÂºÂ¬N THANH TOÃƒÂN THÃƒâ‚¬NH CÃƒâ€NG (Webhook xÃ¡Â»Â­ lÃƒÂ½ sau khi cÃ¡Â»â€¢ng TT gÃ¡Â»Âi vÃƒÂ o)
// CÃ¡ÂºÂ­p nhÃ¡ÂºÂ­t appointment -> CONFIRMED, slot -> OCCUPIED vÃƒÂ  giÃ¡ÂºÂ£i phÃƒÂ³ng lock
// =====================================================================
func (r *pgRepository) ConfirmPayment(appointmentID string) error {
	return r.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultSuccess,
	})
}

// HandlePaymentResult is retained for repository interface compatibility.
func (r *pgRepository) HandlePaymentResult(command HandlePaymentResultCommand) error {
	return appappointment.NewUsecase(r).HandlePaymentResult(command)
}
