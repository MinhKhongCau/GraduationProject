package appointment

import (
	"testing"
	"time"
)

func TestMeetingWindowStatus(t *testing.T) {
	window := MeetingWindow{Early: 15 * time.Minute, Late: 60 * time.Minute}
	start := int64(10 * 60 * 60 * 1000)
	end := start + int64(time.Hour/time.Millisecond)
	minute := int64(time.Minute / time.Millisecond)
	confirmed := Appointment{Status: AppointmentStatusConfirmed, StartTime: start, EndTime: end}
	completedAt := end

	cases := []struct {
		name string
		appt Appointment
		now  int64
		want string
	}{
		{"before early window", confirmed, start - 16*minute, MeetingStatusTooEarly},
		{"early window opens", confirmed, start - 15*minute, MeetingStatusOpen},
		{"during session", confirmed, start + 30*minute, MeetingStatusOpen},
		{"late window last ms", confirmed, end + 60*minute, MeetingStatusOpen},
		{"after late window", confirmed, end + 61*minute, MeetingStatusTooLate},
		{"pending payment has no room", Appointment{Status: AppointmentStatusPendingPayment, StartTime: start, EndTime: end}, start, ""},
		{"completed has no room", Appointment{Status: AppointmentStatusCompleted, StartTime: start, EndTime: end}, start, ""},
		{"session already ended", Appointment{Status: AppointmentStatusConfirmed, StartTime: start, EndTime: end, SessionCompletedAt: &completedAt}, start, ""},
		{"slot time unknown", Appointment{Status: AppointmentStatusConfirmed}, start, ""},
	}
	for _, tc := range cases {
		if got := window.MeetingStatus(tc.appt, tc.now); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestFillSlot(t *testing.T) {
	a := Appointment{SlotID: "s1", StartTime: 1_000, EndTime: 1_000 + 45*60_000, Price: 200000}
	a.FillSlot()
	if a.Slot == nil || a.Slot.DurationMinutes != 45 || a.Slot.Price != 200000 || a.Slot.SlotID != "s1" {
		t.Fatalf("unexpected slot: %+v", a.Slot)
	}
	empty := Appointment{}
	empty.FillSlot()
	if empty.Slot != nil {
		t.Fatalf("slot should stay nil without slot time")
	}
}
