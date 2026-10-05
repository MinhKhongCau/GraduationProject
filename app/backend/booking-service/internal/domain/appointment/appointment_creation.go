package appointment

import (
	"booking-service/internal/domain/shared"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
)

var (
	ErrAppointmentCreationSlotNotLockedByPatient = errors.New("slot is not locked by patient")
	ErrAppointmentCreationSlotExpertMismatch     = errors.New("slot expert does not match appointment expert")
	ErrAppointmentCreationSlotLockExpired        = errors.New("slot lock expired")
)

func ValidateAppointmentCreationSlot(slot slotdomain.ExpertSlot, appointment *Appointment, nowMs int64) error {
	if slot.Status != slotdomain.SlotStatusLocked || slot.LockedBy == nil || !shared.SameID(*slot.LockedBy, appointment.PatientID) {
		return ErrAppointmentCreationSlotNotLockedByPatient
	}
	if !shared.SameID(slot.ExpertID, appointment.ExpertID) {
		return ErrAppointmentCreationSlotExpertMismatch
	}
	if slot.LockedExpiresAt == nil {
		return ErrAppointmentCreationSlotLockExpired
	}
	if *slot.LockedExpiresAt <= nowMs {
		return ErrAppointmentCreationSlotLockExpired
	}
	return nil
}
