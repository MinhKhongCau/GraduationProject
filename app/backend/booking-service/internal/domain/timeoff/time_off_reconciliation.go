package timeoff

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
)

var ErrTimeOffProtectedConflict = errors.New("time-off conflicts with protected booking")

type TimeOffReconciliationPlan struct {
	ConflictIDs            []string
	CancelAppointmentIDs   []string
	MakeUnavailableSlotIDs []string
}

func PlanTimeOffReconciliation(slots []slotdomain.ExpertSlot, appointments []appointmentdomain.Appointment, force bool, nowMs int64) (TimeOffReconciliationPlan, error) {
	appointmentBySlot := make(map[string]appointmentdomain.Appointment, len(appointments))
	for _, appointment := range appointments {
		appointmentBySlot[appointment.SlotID] = appointment
	}
	plan := TimeOffReconciliationPlan{}
	for _, slot := range slots {
		appointment, hasAppointment := appointmentBySlot[slot.SlotID]
		conflictID := slot.SlotID
		if hasAppointment {
			conflictID = appointment.AppointmentID
		}
		if slot.Status == slotdomain.SlotStatusOccupied || (hasAppointment && appointment.Status == appointmentdomain.AppointmentStatusConfirmed) {
			plan.ConflictIDs = append(plan.ConflictIDs, conflictID)
			continue
		}
		activeLock := slot.Status == slotdomain.SlotStatusLocked && (slot.LockedExpiresAt == nil || *slot.LockedExpiresAt > nowMs)
		pending := hasAppointment && appointment.Status == appointmentdomain.AppointmentStatusPendingPayment
		if (!force && (activeLock || pending)) || (force && activeLock && !pending) {
			plan.ConflictIDs = append(plan.ConflictIDs, conflictID)
			continue
		}
		if force && pending {
			plan.CancelAppointmentIDs = append(plan.CancelAppointmentIDs, appointment.AppointmentID)
		}
		if slot.Status == slotdomain.SlotStatusAvailable || slot.Status == slotdomain.SlotStatusLocked {
			plan.MakeUnavailableSlotIDs = append(plan.MakeUnavailableSlotIDs, slot.SlotID)
		}
	}
	if len(plan.ConflictIDs) > 0 {
		return plan, ErrTimeOffProtectedConflict
	}
	return plan, nil
}
