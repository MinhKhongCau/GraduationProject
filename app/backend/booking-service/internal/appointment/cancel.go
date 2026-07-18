package appointment

func (u *appointmentUsecase) CancelAppointment(appointmentID, userID, userRole, reason string) error {
	if userRole == "PATIENT" {
		// TODO: Má»‘c thá»i gian huá»· tá»‘i thiá»ƒu (policy)
		// Check náº¿u cuá»™c háº¹n báº¯t Ä‘áº§u trong vÃ²ng 24h thÃ¬ cháº·n khÃ´ng cho huá»·
		return u.repo.CancelAppointmentByPatient(appointmentID, userID, reason)
	} else if userRole == "EXPERT" {
		return u.repo.CancelAppointmentByExpert(appointmentID, reason)
	}
	return ErrUnauthorized
}
