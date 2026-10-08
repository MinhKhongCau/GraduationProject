package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeProfileDirectory struct {
	profiles map[uuid.UUID]PartySummary
	err      error
	calls    int
}

func (f *fakeProfileDirectory) GetProfileSummaries(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]PartySummary, error) {
	f.calls++
	return f.profiles, f.err
}

type fakeAppointmentDirectory struct {
	appointments map[uuid.UUID]AppointmentInfo
	err          error
}

func (f *fakeAppointmentDirectory) GetAppointmentSummaries(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]AppointmentInfo, error) {
	return f.appointments, f.err
}

func TestOrderViewsIncludeNamesAndAppointmentTime(t *testing.T) {
	f := newManageFixture(t)
	profiles := &fakeProfileDirectory{profiles: map[uuid.UUID]PartySummary{
		f.patient: {ID: f.patient, FullName: "Patient P"},
		f.expertA: {ID: f.expertA, FullName: "Dr. A"},
	}}
	appointments := &fakeAppointmentDirectory{appointments: map[uuid.UUID]AppointmentInfo{
		*f.paidA.AppointmentID: {ID: *f.paidA.AppointmentID, Status: "CONFIRMED", StartTime: 111, SpecializationName: "Tâm lý"},
	}}
	usecase := NewUsecaseWithOptions(f.repo, f.repo, &fakePaymentGateway{}, &fakeBookingClient{}, Options{
		Clock: time.Now, ManagedExperts: f.repo, Profiles: profiles, Appointments: appointments,
	})
	manager := usecase.(ManageUsecase)

	page, err := manager.ListAdminOrders(context.Background(), f.adminA, f.filter)
	if err != nil {
		t.Fatal(err)
	}
	if profiles.calls != 1 {
		t.Fatalf("profiles must be looked up once per page, got %d calls", profiles.calls)
	}
	var paid *AdminPaymentOrderView
	for i := range page.Items {
		if page.Items[i].ID == f.paidA.ID {
			paid = &page.Items[i]
		}
	}
	if paid == nil || paid.Payer == nil || paid.Payer.FullName != "Patient P" || paid.Expert == nil || paid.Expert.FullName != "Dr. A" ||
		paid.Appointment == nil || paid.Appointment.StartTime != 111 {
		t.Fatalf("admin view not enriched: %+v", paid)
	}

	detail, err := manager.GetExpertOrder(context.Background(), f.expertA, f.paidA.ID)
	if err != nil || detail.Payer == nil || detail.Appointment == nil || detail.Appointment.SpecializationName != "Tâm lý" {
		t.Fatalf("expert detail not enriched: %+v err=%v", detail, err)
	}

	profiles.profiles, profiles.err = nil, errors.New("profile-service down")
	appointments.appointments, appointments.err = nil, errors.New("booking-service down")
	page, err = manager.ListAdminOrders(context.Background(), f.adminA, f.filter)
	if err != nil || page.TotalItems != 2 || page.Items[0].Payer != nil {
		t.Fatalf("lookup failures must degrade to plain orders, got %+v err=%v", page, err)
	}
}
