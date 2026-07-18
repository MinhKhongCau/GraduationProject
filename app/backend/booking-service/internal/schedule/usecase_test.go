package schedule

import (
	"booking-service/internal/booking/domain"
	"errors"
	"testing"
	"time"
)

type fakeScheduleRepo struct {
	templates      map[string]domain.TimeTemplate
	availabilities map[string]domain.Availability
	created        *domain.Availability
	updated        map[string]interface{}
}

func newFakeScheduleRepo() *fakeScheduleRepo {
	return &fakeScheduleRepo{templates: map[string]domain.TimeTemplate{}, availabilities: map[string]domain.Availability{}}
}
func (r *fakeScheduleRepo) GetAvailabilities(expertID string) ([]domain.Availability, error) {
	var out []domain.Availability
	for _, a := range r.availabilities {
		if a.ExpertID == expertID && a.IsEnabled {
			out = append(out, a)
		}
	}
	return out, nil
}
func (r *fakeScheduleRepo) GetTimeTemplates() ([]domain.TimeTemplate, error) {
	var out []domain.TimeTemplate
	for _, template := range r.templates {
		if template.IsActive {
			out = append(out, template)
		}
	}
	return out, nil
}
func (r *fakeScheduleRepo) GetAllTimeTemplates() ([]domain.TimeTemplate, error) {
	var out []domain.TimeTemplate
	for _, template := range r.templates {
		out = append(out, template)
	}
	return out, nil
}
func (r *fakeScheduleRepo) GetTimeTemplateByID(id string) (*domain.TimeTemplate, error) {
	template, ok := r.templates[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &template, nil
}
func (r *fakeScheduleRepo) GetAvailabilityByID(id, expertID string) (*domain.Availability, error) {
	availability, ok := r.availabilities[id]
	if !ok || availability.ExpertID != expertID {
		return nil, ErrNotFound
	}
	return &availability, nil
}
func (r *fakeScheduleRepo) GetEnabledAvailabilities(expertID string) ([]domain.Availability, error) {
	return r.GetAvailabilities(expertID)
}
func (r *fakeScheduleRepo) GetEnabledAvailabilitiesByTemplate(id string) ([]domain.Availability, error) {
	var out []domain.Availability
	for _, a := range r.availabilities {
		if a.TemplateID == id && a.IsEnabled {
			out = append(out, a)
		}
	}
	return out, nil
}
func (r *fakeScheduleRepo) CreateTimeTemplate(template *domain.TimeTemplate) error {
	r.templates[template.TemplateID] = *template
	return nil
}
func (r *fakeScheduleRepo) CreateAvailability(availability *domain.Availability) error {
	copy := *availability
	r.created = &copy
	r.availabilities[availability.AvailabilityID] = copy
	return nil
}
func (r *fakeScheduleRepo) UpdateAvailability(id, expertID string, updates map[string]interface{}) error {
	r.updated = updates
	return nil
}
func (r *fakeScheduleRepo) UpdateTemplate(id string, updates map[string]interface{}) error {
	r.updated = updates
	return nil
}

func TestCreateAvailabilityValidatesTemplateWeekdayRangeAndPrice(t *testing.T) {
	repo := newFakeScheduleRepo()
	repo.templates["active"] = testTemplate("active", "08:00", "09:00", true)
	u := NewUsecase(repo)
	if _, err := u.CreateAvailability("expert", "missing", 1, 1, nil, 200000); !errors.Is(err, ErrScheduleNotFound) {
		t.Fatalf("missing template: %v", err)
	}
	for _, weekday := range []int{0, 8} {
		if _, err := u.CreateAvailability("expert", "active", weekday, 1, nil, 200000); !errors.Is(err, ErrInvalidSchedule) {
			t.Fatalf("weekday %d: %v", weekday, err)
		}
	}
	inactive := testTemplate("inactive", "08:00", "09:00", false)
	repo.templates["inactive"] = inactive
	if _, err := u.CreateAvailability("expert", "inactive", 1, 1, nil, 200000); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("inactive template: %v", err)
	}
	if _, err := u.CreateAvailability("expert", "active", 1, 1, int64Pointer(0), 200000); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("invalid effective range: %v", err)
	}
	availability, err := u.CreateAvailability("expert", "active", 1, 1, nil, 200000)
	if err != nil {
		t.Fatal(err)
	}
	if availability.Price == nil || *availability.Price != 200000 || repo.created == nil {
		t.Fatal("validated price was not persisted")
	}
}

func TestUpdateAvailabilityValidatesOwnershipAndPatchedValues(t *testing.T) {
	repo := newFakeScheduleRepo()
	repo.templates["tpl"] = testTemplate("tpl", "08:00", "09:00", true)
	price := 200000.0
	repo.availabilities["avail"] = domain.Availability{AvailabilityID: "avail", ExpertID: "expert", TemplateID: "tpl", DayOfWeek: 1, IsEnabled: true, EffectiveFrom: 100, Price: &price}
	u := NewUsecase(repo)
	if err := u.UpdateAvailability("avail", "other", map[string]interface{}{"day_of_week": 2}); !errors.Is(err, ErrScheduleNotFound) {
		t.Fatalf("ownership must be enforced: %v", err)
	}
	if err := u.UpdateAvailability("avail", "expert", map[string]interface{}{"day_of_week": 8}); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("invalid patched weekday: %v", err)
	}
	if err := u.UpdateAvailability("avail", "expert", map[string]interface{}{"effective_until": int64(99)}); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("invalid patched range: %v", err)
	}
	if err := u.UpdateAvailability("avail", "expert", map[string]interface{}{"template_id": "missing"}); !errors.Is(err, ErrScheduleNotFound) {
		t.Fatalf("missing patched template: %v", err)
	}
	repo.templates["inactive"] = testTemplate("inactive", "08:00", "09:00", false)
	if err := u.UpdateAvailability("avail", "expert", map[string]interface{}{"template_id": "inactive", "is_enabled": true}); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("enabling with inactive template: %v", err)
	}
	if err := u.UpdateAvailability("avail", "expert", map[string]interface{}{"price": float64(250000)}); err != nil {
		t.Fatal(err)
	}
	if repo.updated["price"] != float64(250000) {
		t.Fatal("valid patch was not persisted")
	}
}

func TestAvailabilityOverlapValidationUsesTimeAndEffectiveRanges(t *testing.T) {
	repo := newFakeScheduleRepo()
	repo.templates["morning"] = testTemplate("morning", "08:00", "09:00", true)
	repo.templates["overlap"] = testTemplate("overlap", "08:30", "09:30", true)
	price := 200000.0
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, businessLocation).UnixMilli()
	until := time.Date(2026, 7, 31, 0, 0, 0, 0, businessLocation).UnixMilli()
	repo.availabilities["existing"] = domain.Availability{AvailabilityID: "existing", ExpertID: "expert", TemplateID: "morning", DayOfWeek: 1, IsEnabled: true, EffectiveFrom: from, EffectiveUntil: &until, Price: &price}
	u := NewUsecase(repo)
	if _, err := u.CreateAvailability("expert", "overlap", 1, time.Date(2026, 7, 15, 0, 0, 0, 0, businessLocation).UnixMilli(), nil, 200000); !errors.Is(err, ErrScheduleOverlap) {
		t.Fatalf("expected overlapping effective schedules to fail: %v", err)
	}
	if _, err := u.CreateAvailability("expert", "overlap", 1, time.Date(2026, 8, 1, 0, 0, 0, 0, businessLocation).UnixMilli(), nil, 200000); err != nil {
		t.Fatalf("non-overlapping effective range should pass: %v", err)
	}
}

func TestCreateTimeTemplateUsesStrictDomainValidation(t *testing.T) {
	u := NewUsecase(newFakeScheduleRepo())
	for _, times := range [][2]string{{"8:00", "09:00"}, {"09:00", "09:00"}, {"10:00", "09:00"}} {
		if _, err := u.CreateTimeTemplate("shift", times[0], times[1], 30); !errors.Is(err, ErrInvalidSchedule) {
			t.Fatalf("expected %v to fail: %v", times, err)
		}
	}
	if _, err := u.CreateTimeTemplate("shift", "08:00", "09:00", 0); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("zero duration: %v", err)
	}
}

func testTemplate(id, start, end string, active bool) domain.TimeTemplate {
	return domain.TimeTemplate{TemplateID: id, StartTime: start, EndTime: end, SlotDurationMinutes: 30, IsActive: active}
}
func int64Pointer(value int64) *int64 { return &value }
