package domain

import (
	"errors"
	"testing"
)

func TestNormalTimeOffProtectedConflictRules(t *testing.T) {
	now := int64(100)
	future := int64(200)
	past := int64(50)
	for _, tt := range []struct {
		name        string
		slot        ExpertSlot
		appointment *Appointment
		conflict    bool
	}{
		{"available", ExpertSlot{SlotID: "slot", Status: SlotStatusAvailable}, nil, false},
		{"expired standalone lock", ExpertSlot{SlotID: "slot", Status: SlotStatusLocked, LockedExpiresAt: &past}, nil, false},
		{"active standalone lock", ExpertSlot{SlotID: "slot", Status: SlotStatusLocked, LockedExpiresAt: &future}, nil, true},
		{"pending payment", ExpertSlot{SlotID: "slot", Status: SlotStatusLocked, LockedExpiresAt: &future}, &Appointment{AppointmentID: "appt", SlotID: "slot", Status: AppointmentStatusPendingPayment}, true},
		{"confirmed", ExpertSlot{SlotID: "slot", Status: SlotStatusOccupied}, &Appointment{AppointmentID: "appt", SlotID: "slot", Status: AppointmentStatusConfirmed}, true},
		{"occupied", ExpertSlot{SlotID: "slot", Status: SlotStatusOccupied}, nil, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			appointments := []Appointment(nil)
			if tt.appointment != nil {
				appointments = append(appointments, *tt.appointment)
			}
			plan, err := PlanTimeOffReconciliation([]ExpertSlot{tt.slot}, appointments, false, now)
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
	pendingSlot := ExpertSlot{SlotID: "pending-slot", Status: SlotStatusLocked, LockedExpiresAt: &future}
	pending := Appointment{AppointmentID: "pending", SlotID: pendingSlot.SlotID, Status: AppointmentStatusPendingPayment}
	plan, err := PlanTimeOffReconciliation([]ExpertSlot{pendingSlot}, []Appointment{pending}, true, 100)
	if err != nil || len(plan.CancelAppointmentIDs) != 1 || plan.CancelAppointmentIDs[0] != pending.AppointmentID || len(plan.MakeUnavailableSlotIDs) != 1 {
		t.Fatalf("unexpected force plan: %+v err=%v", plan, err)
	}

	for _, tt := range []struct {
		name string
		slot ExpertSlot
		appt *Appointment
	}{
		{"confirmed", ExpertSlot{SlotID: "confirmed-slot", Status: SlotStatusOccupied}, &Appointment{AppointmentID: "confirmed", SlotID: "confirmed-slot", Status: AppointmentStatusConfirmed}},
		{"occupied", ExpertSlot{SlotID: "occupied-slot", Status: SlotStatusOccupied}, nil},
		{"active standalone lock", ExpertSlot{SlotID: "locked-slot", Status: SlotStatusLocked, LockedExpiresAt: &future}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			appointments := []Appointment(nil)
			if tt.appt != nil {
				appointments = append(appointments, *tt.appt)
			}
			plan, err := PlanTimeOffReconciliation([]ExpertSlot{tt.slot}, appointments, true, 100)
			if !errors.Is(err, ErrTimeOffProtectedConflict) || len(plan.CancelAppointmentIDs) != 0 {
				t.Fatalf("protected booking was not preserved: %+v err=%v", plan, err)
			}
		})
	}
}
