package domain

import (
	"errors"
)

var (
	ErrAppointmentCreationSlotNotLockedByPatient = errors.New("slot is not locked by patient")
	ErrAppointmentCreationSlotExpertMismatch     = errors.New("slot expert does not match appointment expert")
	ErrAppointmentCreationSlotLockExpired        = errors.New("slot lock expired")
)

func ValidateAppointmentCreationSlot(slot ExpertSlot, appointment *Appointment, nowMs int64) error {
	if slot.Status != SlotStatusLocked || slot.LockedBy == nil || !SameID(*slot.LockedBy, appointment.PatientID) {
		return ErrAppointmentCreationSlotNotLockedByPatient
	}
	if !SameID(slot.ExpertID, appointment.ExpertID) {
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
