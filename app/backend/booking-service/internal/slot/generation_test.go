package slot

import (
	"booking-service/internal/booking/domain"
	"errors"
	"math"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fakeAppointmentLocker struct{}

func (fakeAppointmentLocker) LockSlot(string, string) error { return nil }

type generationRepo struct {
	existing  []domain.ExpertSlot
	inserted  []domain.ExpertSlot
	bulkCalls int
	err       error
}

func (r *generationRepo) GetAvailableDates(string, time.Time, time.Time) ([]string, error) {
	return nil, nil
}
func (r *generationRepo) GetAvailableTimes(string, string) ([]SlotTimeResult, error) { return nil, nil }
func (r *generationRepo) GetSlotsByExpert(string, int64, int64) ([]domain.ExpertSlot, error) {
	return nil, nil
}
func (r *generationRepo) GetOverlappingSlots(string, int64, int64) ([]domain.ExpertSlot, error) {
	return r.existing, r.err
}
func (r *generationRepo) BulkInsertSlots(slots []domain.ExpertSlot) (int64, error) {
	r.bulkCalls++
	r.inserted = append(r.inserted, slots...)
	return int64(len(slots)), r.err
}

func TestGeneratedSlotSnapshotsValidatedAvailabilityPrice(t *testing.T) {
	now := generationNow()
	repo := &generationRepo{}
	u := NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now: now})
	result, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{generationAvailability(now, floatPointer(250000))}, []domain.TimeTemplate{generationTemplate("08:30", "09:30", 30)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Inserted != 2 || len(repo.inserted) != 2 {
		t.Fatalf("expected two inserts, got %#v", result)
	}
	for _, slot := range repo.inserted {
		if slot.Price != 250000 {
			t.Fatalf("expected price snapshot 250000, got %v", slot.Price)
		}
	}
}

func TestGenerationRejectsInvalidPricesBeforePersistence(t *testing.T) {
	prices := []*float64{nil, floatPointer(0), floatPointer(-1), floatPointer(1.5), floatPointer(math.NaN()), floatPointer(math.Inf(1)), floatPointer(math.Inf(-1)), floatPointer(92233720368547760)}
	for index, price := range prices {
		t.Run(time.Duration(index).String(), func(t *testing.T) {
			now := generationNow()
			repo := &generationRepo{}
			u := NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
			_, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{generationAvailability(now, price)}, []domain.TimeTemplate{generationTemplate("08:30", "09:30", 30)}, nil)
			if !errors.Is(err, ErrInvalidGeneration) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if repo.bulkCalls != 0 {
				t.Fatal("invalid price must not persist slots")
			}
		})
	}
}

func TestGenerationHonorsEffectiveRangeEnabledAndTemplateStatus(t *testing.T) {
	now := generationNow()
	tests := []struct {
		name   string
		mutate func(*domain.Availability, *domain.TimeTemplate)
		want   int64
	}{
		{"before effective from", func(a *domain.Availability, _ *domain.TimeTemplate) {
			a.EffectiveFrom = now.AddDate(0, 0, 1).UnixMilli()
		}, 0},
		{"equal effective from", func(a *domain.Availability, _ *domain.TimeTemplate) {
			a.EffectiveFrom = now.Add(12 * time.Hour).UnixMilli()
		}, 2},
		{"equal effective until", func(a *domain.Availability, _ *domain.TimeTemplate) {
			until := now.Add(-time.Hour).UnixMilli()
			a.EffectiveUntil = &until
		}, 2},
		{"after effective until", func(a *domain.Availability, _ *domain.TimeTemplate) {
			until := now.AddDate(0, 0, -1).UnixMilli()
			a.EffectiveUntil = &until
		}, 0},
		{"open ended", func(a *domain.Availability, _ *domain.TimeTemplate) { a.EffectiveUntil = nil }, 2},
		{"disabled", func(a *domain.Availability, _ *domain.TimeTemplate) { a.IsEnabled = false }, 0},
		{"inactive template", func(_ *domain.Availability, tpl *domain.TimeTemplate) { tpl.IsActive = false }, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := generationAvailability(now, floatPointer(200000))
			tpl := generationTemplate("08:30", "09:30", 30)
			tt.mutate(&a, &tpl)
			repo := &generationRepo{}
			u := NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
			result, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{a}, []domain.TimeTemplate{tpl}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.Inserted != tt.want {
				t.Fatalf("got %d inserts, want %d", result.Inserted, tt.want)
			}
		})
	}
}

func TestGenerationCutoffAndTrailingRemainder(t *testing.T) {
	now := generationNow()
	tests := []struct {
		name, start, end string
		duration         int
		wantStarts       []string
	}{
		{"past", "07:00", "08:00", 30, nil},
		{"exact cutoff", "08:05", "08:35", 30, nil},
		{"after cutoff", "08:06", "08:36", 30, []string{"08:06"}},
		{"mixed same day", "07:30", "09:30", 30, []string{"08:30", "09:00"}},
		{"trailing remainder", "08:10", "10:20", 30, []string{"08:10", "08:40", "09:10", "09:40"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &generationRepo{}
			u := NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
			_, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{generationAvailability(now, floatPointer(200000))}, []domain.TimeTemplate{generationTemplate(tt.start, tt.end, tt.duration)}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(repo.inserted) != len(tt.wantStarts) {
				t.Fatalf("got %d slots, want %d", len(repo.inserted), len(tt.wantStarts))
			}
			for i, want := range tt.wantStarts {
				if got := time.UnixMilli(repo.inserted[i].StartTime).In(generationLocation).Format("15:04"); got != want {
					t.Fatalf("slot %d starts %s, want %s", i, got, want)
				}
			}
		})
	}
}

func TestGenerationFiltersTimeOffPerSlotWithHalfOpenIntervals(t *testing.T) {
	now := generationNow()
	baseStart := time.Date(2026, 7, 20, 8, 30, 0, 0, generationLocation)
	tests := []struct {
		name             string
		offStart, offEnd time.Time
		want             int64
	}{
		{"non-overlap", baseStart.Add(2 * time.Hour), baseStart.Add(3 * time.Hour), 2},
		{"partial overlap", baseStart.Add(15 * time.Minute), baseStart.Add(45 * time.Minute), 0},
		{"slot ends at timeoff start", baseStart.Add(30 * time.Minute), baseStart.Add(45 * time.Minute), 1},
		{"slot starts at timeoff end", baseStart.Add(-15 * time.Minute), baseStart, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &generationRepo{}
			u := NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
			off := domain.ExpertTimeOff{StartDatetime: tt.offStart.UnixMilli(), EndDatetime: tt.offEnd.UnixMilli()}
			result, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{generationAvailability(now, floatPointer(200000))}, []domain.TimeTemplate{generationTemplate("08:30", "09:30", 30)}, []domain.ExpertTimeOff{off})
			if err != nil {
				t.Fatal(err)
			}
			if result.Inserted != tt.want {
				t.Fatalf("got %d inserts, want %d", result.Inserted, tt.want)
			}
		})
	}
}

func TestOneHourTimeOffDoesNotSuppressUnrelatedSlotsForWholeDay(t *testing.T) {
	now := generationNow()
	repo := &generationRepo{}
	u := NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
	offStart := time.Date(2026, 7, 20, 9, 0, 0, 0, generationLocation)
	off := domain.ExpertTimeOff{StartDatetime: offStart.UnixMilli(), EndDatetime: offStart.Add(time.Hour).UnixMilli()}
	result, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{generationAvailability(now, floatPointer(200000))}, []domain.TimeTemplate{generationTemplate("08:30", "12:30", 30)}, []domain.ExpertTimeOff{off})
	if err != nil {
		t.Fatal(err)
	}
	if result.Inserted != 6 {
		t.Fatalf("expected six unrelated slots to remain, got %d", result.Inserted)
	}
}

func TestGenerationRejectsCandidateAndPersistedOverlapsButAllowsIdempotency(t *testing.T) {
	now := generationNow()
	availability := generationAvailability(now, floatPointer(200000))
	second := availability
	second.AvailabilityID = "avail-2"
	second.TemplateID = "tpl-2"
	repo := &generationRepo{}
	u := NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
	_, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{availability, second}, []domain.TimeTemplate{generationTemplate("08:30", "09:30", 30), {TemplateID: "tpl-2", StartTime: "08:45", EndTime: "09:45", SlotDurationMinutes: 30, IsActive: true}}, nil)
	if !errors.Is(err, ErrSlotOverlap) || repo.bulkCalls != 0 {
		t.Fatalf("expected candidate overlap rejection, got %v", err)
	}

	start := time.Date(2026, 7, 20, 8, 30, 0, 0, generationLocation)
	repo = &generationRepo{existing: []domain.ExpertSlot{{SlotID: "existing", ExpertID: "expert", StartTime: start.UnixMilli(), EndTime: start.Add(30 * time.Minute).UnixMilli(), Price: 100000}}}
	u = NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
	result, err := u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{availability}, []domain.TimeTemplate{generationTemplate("08:30", "09:00", 30)}, nil)
	if err != nil || result.Inserted != 0 || repo.existing[0].Price != 100000 {
		t.Fatalf("exact regeneration should be no-op without repricing: %#v, %v", result, err)
	}

	repo = &generationRepo{existing: []domain.ExpertSlot{{SlotID: "existing", ExpertID: "expert", StartTime: start.Add(15 * time.Minute).UnixMilli(), EndTime: start.Add(45 * time.Minute).UnixMilli()}}}
	u = NewUsecaseWithClock(repo, fakeAppointmentLocker{}, fixedClock{now})
	_, err = u.GenerateSlotsForNextDays("expert", 1, []domain.Availability{availability}, []domain.TimeTemplate{generationTemplate("08:30", "09:00", 30)}, nil)
	if !errors.Is(err, ErrSlotOverlap) || repo.bulkCalls != 0 {
		t.Fatalf("expected persisted overlap rejection, got %v", err)
	}
}

func generationNow() time.Time { return time.Date(2026, 7, 20, 8, 0, 0, 0, generationLocation) }
func generationAvailability(now time.Time, price *float64) domain.Availability {
	return domain.Availability{AvailabilityID: "avail", ExpertID: "expert", TemplateID: "tpl", DayOfWeek: 1, IsEnabled: true, EffectiveFrom: now.Add(-24 * time.Hour).UnixMilli(), Price: price}
}
func generationTemplate(start, end string, duration int) domain.TimeTemplate {
	return domain.TimeTemplate{TemplateID: "tpl", StartTime: start, EndTime: end, SlotDurationMinutes: duration, IsActive: true}
}
func floatPointer(value float64) *float64 { return &value }
