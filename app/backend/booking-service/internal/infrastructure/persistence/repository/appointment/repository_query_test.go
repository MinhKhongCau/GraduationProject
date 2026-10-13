package repository

import (
	appappointment "booking-service/internal/application/appointment"
	"testing"
)

func TestAppointmentOrder(t *testing.T) {
	cases := []struct {
		sort appappointment.AppointmentSort
		want string
	}{
		{appappointment.AppointmentSort{}, "slot.start_time DESC, appointment.appointment_id DESC"},
		{appappointment.AppointmentSort{By: appappointment.AppointmentSortByStartTime}, "slot.start_time DESC, appointment.appointment_id DESC"},
		{appappointment.AppointmentSort{By: appappointment.AppointmentSortByStartTime, Ascending: true}, "slot.start_time ASC, appointment.appointment_id ASC"},
		{appappointment.AppointmentSort{By: appappointment.AppointmentSortByCreatedAt}, "appointment.created_at DESC, appointment.appointment_id DESC"},
	}
	for _, tc := range cases {
		if got := appointmentOrder(tc.sort); got != tc.want {
			t.Fatalf("appointmentOrder(%+v) = %q, want %q", tc.sort, got, tc.want)
		}
	}
}
