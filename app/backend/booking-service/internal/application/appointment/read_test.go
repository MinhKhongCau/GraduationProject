package appointment

import (
	bookingquery "booking-service/internal/application/query"
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"errors"
	"testing"
)

func TestListAppointmentsEnforcesActorOwnershipAndReturnsPage(t *testing.T) {
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, slotdomain.ExpertSlot{})
	repo.readItems = []appointmentdomain.Appointment{{AppointmentID: "appointment-1", PatientID: "patient-1"}}
	repo.readTotal = 3
	usecase := NewUsecase(repo).(ReadUsecase)
	result, err := usecase.ListAppointments(AppointmentListQuery{ActorID: "patient-1", ActorRole: "PATIENT", ExpertID: "expert-filter", Page: bookingquery.PageRequest{Page: 1, Size: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if repo.readFilter.PatientID != "patient-1" || repo.readFilter.ExpertID != "expert-filter" {
		t.Fatalf("ownership filter not enforced: %+v", repo.readFilter)
	}
	if result.TotalPages != 2 || !result.HasPrevious || result.HasNext || len(result.Items) != 1 {
		t.Fatalf("unexpected page: %+v", result)
	}
}

func TestGetAppointmentDetailRejectsUnrelatedActors(t *testing.T) {
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, slotdomain.ExpertSlot{})
	repo.detail = &appointmentdomain.Appointment{AppointmentID: "appointment-1", PatientID: "patient-1", ExpertID: "expert-1"}
	usecase := NewUsecase(repo).(ReadUsecase)
	if _, err := usecase.GetAppointmentDetail("patient-2", "PATIENT", "appointment-1"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected patient ownership error, got %v", err)
	}
	if _, err := usecase.GetAppointmentDetail("expert-2", "EXPERT", "appointment-1"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected expert ownership error, got %v", err)
	}
	if got, err := usecase.GetAppointmentDetail("patient-1", "PATIENT", "appointment-1"); err != nil || got.AppointmentID != "appointment-1" {
		t.Fatalf("owner read failed: %+v %v", got, err)
	}
}

func TestParseAppointmentStatusRejectsUnsupportedValue(t *testing.T) {
	if _, err := ParseAppointmentStatus("UNKNOWN"); !errors.Is(err, bookingquery.ErrInvalidQuery) {
		t.Fatalf("expected invalid query, got %v", err)
	}
}
