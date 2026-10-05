package timeoff

import (
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"testing"
)

func TestNormalTimeOffProtectedConflictRules(t *testing.T) {
	now := int64(100)
	future := int64(200)
	past := int64(50)
	for _, tt := range []struct {
		name        string
		slot        slotdomain.ExpertSlot
		appointment *appointmentdomain.Appointment
		conflict    bool
	}{
		{"available", slotdomain.ExpertSlot{SlotID: "slot", Status: slotdomain.SlotStatusAvailable}, nil, false},
		{"expired standalone lock", slotdomain.ExpertSlot{SlotID: "slot", Status: slotdomain.SlotStatusLocked, LockedExpiresAt: &past}, nil, false},
		{"active standalone lock", slotdomain.ExpertSlot{SlotID: "slot", Status: slotdomain.SlotStatusLocked, LockedExpiresAt: &future}, nil, true},
		{"pending payment", slotdomain.ExpertSlot{SlotID: "slot", Status: slotdomain.SlotStatusLocked, LockedExpiresAt: &future}, &appointmentdomain.Appointment{AppointmentID: "appt", SlotID: "slot", Status: appointmentdomain.AppointmentStatusPendingPayment}, true},
		{"confirmed", slotdomain.ExpertSlot{SlotID: "slot", Status: slotdomain.SlotStatusOccupied}, &appointmentdomain.Appointment{AppointmentID: "appt", SlotID: "slot", Status: appointmentdomain.AppointmentStatusConfirmed}, true},
		{"occupied", slotdomain.ExpertSlot{SlotID: "slot", Status: slotdomain.SlotStatusOccupied}, nil, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			appointments := []appointmentdomain.Appointment(nil)
			if tt.appointment != nil {
				appointments = append(appointments, *tt.appointment)
			}
			plan, err := PlanTimeOffReconciliation([]slotdomain.ExpertSlot{tt.slot}, appointments, false, now)
			if errors.Is(err, ErrTimeOffProtectedConflict) != tt.conflict {
				t.Fatalf("conflict=%v plan=%+v err=%v", tt.conflict, plan, err)
			}
			if !tt.conflict && len(plan.MakeUnavailableSlotIDs) != 1 {
				t.Fatalf("safe covered slot was not made unavailable: %+v", plan)
			}
		})
	}
}

func TestForceTimeOffCancelsOnlyPendingPayment(t *testing.T) {
	future := int64(200)
	pendingSlot := slotdomain.ExpertSlot{SlotID: "pending-slot", Status: slotdomain.SlotStatusLocked, LockedExpiresAt: &future}
	pending := appointmentdomain.Appointment{AppointmentID: "pending", SlotID: pendingSlot.SlotID, Status: appointmentdomain.AppointmentStatusPendingPayment}
	plan, err := PlanTimeOffReconciliation([]slotdomain.ExpertSlot{pendingSlot}, []appointmentdomain.Appointment{pending}, true, 100)
	if err != nil || len(plan.CancelAppointmentIDs) != 1 || plan.CancelAppointmentIDs[0] != pending.AppointmentID || len(plan.MakeUnavailableSlotIDs) != 1 {
		t.Fatalf("unexpected force plan: %+v err=%v", plan, err)
	}

	for _, tt := range []struct {
		name string
		slot slotdomain.ExpertSlot
		appt *appointmentdomain.Appointment
	}{
		{"confirmed", slotdomain.ExpertSlot{SlotID: "confirmed-slot", Status: slotdomain.SlotStatusOccupied}, &appointmentdomain.Appointment{AppointmentID: "confirmed", SlotID: "confirmed-slot", Status: appointmentdomain.AppointmentStatusConfirmed}},
		{"occupied", slotdomain.ExpertSlot{SlotID: "occupied-slot", Status: slotdomain.SlotStatusOccupied}, nil},
		{"active standalone lock", slotdomain.ExpertSlot{SlotID: "locked-slot", Status: slotdomain.SlotStatusLocked, LockedExpiresAt: &future}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			appointments := []appointmentdomain.Appointment(nil)
			if tt.appt != nil {
				appointments = append(appointments, *tt.appt)
			}
			plan, err := PlanTimeOffReconciliation([]slotdomain.ExpertSlot{tt.slot}, appointments, true, 100)
			if !errors.Is(err, ErrTimeOffProtectedConflict) || len(plan.CancelAppointmentIDs) != 0 {
				t.Fatalf("protected booking was not preserved: %+v err=%v", plan, err)
			}
		})
	}
}
