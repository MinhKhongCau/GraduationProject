package appointment

import (
	bookingquery "booking-service/internal/application/query"
	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
	"context"
	"errors"
	"testing"
)

func TestListAppointmentsEnforcesActorOwnershipAndReturnsPage(t *testing.T) {
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, slotdomain.ExpertSlot{})
	repo.readItems = []appointmentdomain.Appointment{{AppointmentID: "appointment-1", PatientID: "patient-1"}}
	repo.readTotal = 3
	usecase := NewUsecase(repo).(ReadUsecase)
	result, err := usecase.ListAppointments(context.Background(), AppointmentListQuery{ActorID: "patient-1", ActorRole: "PATIENT", ExpertID: "expert-filter", Page: bookingquery.PageRequest{Page: 1, Size: 2}})
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
	if _, err := usecase.GetAppointmentDetail(context.Background(), "patient-2", "PATIENT", "appointment-1"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected patient ownership error, got %v", err)
	}
	if _, err := usecase.GetAppointmentDetail(context.Background(), "expert-2", "EXPERT", "appointment-1"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected expert ownership error, got %v", err)
	}
	if got, err := usecase.GetAppointmentDetail(context.Background(), "patient-1", "PATIENT", "appointment-1"); err != nil || got.AppointmentID != "appointment-1" {
		t.Fatalf("owner read failed: %+v %v", got, err)
	}
}

func TestParseAppointmentStatusRejectsUnsupportedValue(t *testing.T) {
	if _, err := ParseAppointmentStatus("UNKNOWN"); !errors.Is(err, bookingquery.ErrInvalidQuery) {
		t.Fatalf("expected invalid query, got %v", err)
	}
}

// fakeDirectory giả lập profile-service: hồ sơ theo auth id + chuyên gia mỗi Admin quản lý.
type fakeDirectory struct {
	fakeProfileGateway
	profiles   map[string]appointmentdomain.ParticipantProfile
	managed    map[string][]string
	lookupErr  error
	lookedUpID []string
}

func (d *fakeDirectory) GetProfileSummaries(_ context.Context, ids []string) (map[string]appointmentdomain.ParticipantProfile, error) {
	d.lookedUpID = ids
	return d.profiles, d.lookupErr
}

func (d *fakeDirectory) ListManagedExpertIDs(_ context.Context, adminID string) ([]string, error) {
	return d.managed[adminID], nil
}

func newDirectoryUsecase(repo Repository, directory *fakeDirectory) ReadUsecase {
	return NewUsecaseWithProfiles(repo, directory).(ReadUsecase)
}

func TestListAppointmentsAttachesExpertAndPatientProfiles(t *testing.T) {
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, slotdomain.ExpertSlot{})
	repo.readItems = []appointmentdomain.Appointment{{AppointmentID: "a1", PatientID: "patient-1", ExpertID: "expert-1"}}
	directory := &fakeDirectory{profiles: map[string]appointmentdomain.ParticipantProfile{
		"expert-1":  {AuthID: "expert-1", FullName: "Dr. B"},
		"patient-1": {AuthID: "patient-1", FullName: "Patient P"},
	}}
	page, err := newDirectoryUsecase(repo, directory).ListAppointments(context.Background(), AppointmentListQuery{ActorID: "patient-1", ActorRole: "PATIENT"})
	if err != nil {
		t.Fatal(err)
	}
	item := page.Items[0]
	if item.Expert == nil || item.Expert.FullName != "Dr. B" || item.PatientAccount == nil || item.PatientAccount.FullName != "Patient P" {
		t.Fatalf("profiles not attached: %+v", item)
	}
	if len(directory.lookedUpID) != 2 {
		t.Fatalf("expected one batched lookup with 2 ids, got %v", directory.lookedUpID)
	}

	// The fake repository returns its own slice; a real query returns fresh rows each time.
	repo.readItems = []appointmentdomain.Appointment{{AppointmentID: "a1", PatientID: "patient-1", ExpertID: "expert-1"}}
	directory.lookupErr = errors.New("profile down")
	page, err = newDirectoryUsecase(repo, directory).ListAppointments(context.Background(), AppointmentListQuery{ActorID: "patient-1", ActorRole: "PATIENT"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Expert != nil {
		t.Fatalf("lookup failure must degrade to bare appointments, got %+v err=%v", page.Items, err)
	}
}

func TestAdminSeesOnlyManagedExpertAppointments(t *testing.T) {
	repo := newUOWTestRepo(appointmentdomain.Appointment{}, slotdomain.ExpertSlot{})
	repo.readItems = []appointmentdomain.Appointment{{AppointmentID: "a1", ExpertID: "expert-a"}}
	directory := &fakeDirectory{managed: map[string][]string{"admin-a": {"expert-a"}}}
	usecase := newDirectoryUsecase(repo, directory)
	ctx := context.Background()

	if _, err := usecase.ListAdminAppointments(ctx, AppointmentListQuery{ActorID: "admin-a", ActorRole: "ADMIN"}); err != nil {
		t.Fatal(err)
	}
	if !repo.readFilter.ScopeExperts || len(repo.readFilter.ExpertIDs) != 1 || repo.readFilter.ExpertIDs[0] != "expert-a" {
		t.Fatalf("admin scope not applied: %+v", repo.readFilter)
	}
	if _, err := usecase.ListAdminAppointments(ctx, AppointmentListQuery{ActorID: "admin-a", ActorRole: "ADMIN", ExpertID: "expert-b"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("filtering by an unmanaged expert must be forbidden, got %v", err)
	}
	empty, err := usecase.ListAdminAppointments(ctx, AppointmentListQuery{ActorID: "admin-x", ActorRole: "ADMIN", Page: bookingquery.PageRequest{Size: 20}})
	if err != nil || empty.TotalItems != 0 || len(empty.Items) != 0 {
		t.Fatalf("admin without experts must get an empty page, got %+v err=%v", empty, err)
	}
	if _, err := usecase.ListAdminAppointments(ctx, AppointmentListQuery{ActorID: "patient-1", ActorRole: "PATIENT"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("non-admin must be rejected, got %v", err)
	}

	repo.detail = &appointmentdomain.Appointment{AppointmentID: "a1", ExpertID: "expert-a", PatientID: "patient-1"}
	directory.profiles = map[string]appointmentdomain.ParticipantProfile{
		"expert-a":  {AuthID: "expert-a", Role: "EXPERT", FullName: "Dr. A"},
		"patient-1": {AuthID: "patient-1", Role: "PATIENT", FullName: "Booker B"},
	}
	detail, err := usecase.GetAppointmentDetail(ctx, "admin-a", "ADMIN", "a1")
	if err != nil {
		t.Fatalf("managing admin must read detail: %v", err)
	}
	if detail.Expert == nil || detail.Expert.FullName != "Dr. A" || detail.PatientAccount == nil || detail.PatientAccount.FullName != "Booker B" {
		t.Fatalf("detail must carry expert and booking account from profile-service: %+v", detail)
	}
	if _, err := usecase.GetAppointmentDetail(ctx, "admin-b", "ADMIN", "a1"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("other admin must be forbidden, got %v", err)
	}
	if _, err := NewUsecase(repo).(ReadUsecase).GetAppointmentDetail(ctx, "admin-a", "ADMIN", "a1"); !errors.Is(err, ErrProfileServiceUnavailable) {
		t.Fatalf("admin read without profile directory must be unavailable, got %v", err)
	}
}
