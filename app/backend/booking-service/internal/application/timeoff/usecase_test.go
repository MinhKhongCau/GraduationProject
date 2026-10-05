package timeoff

import (
	timeoffdomain "booking-service/internal/domain/timeoff"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeTimeOffRepository struct {
	created       *timeoffdomain.ExpertTimeOff
	force         bool
	createCalls   int
	createErr     error
	conflicts     []string
	unprocessed   []timeoffdomain.ExpertTimeOff
	processErrors map[string]error
	processed     []string
	getCalls      atomic.Int64
}

func (r *fakeTimeOffRepository) CreateAndProcessTimeOff(item *timeoffdomain.ExpertTimeOff, force bool, _ int64) ([]string, error) {
	r.createCalls++
	r.created, r.force = item, force
	return r.conflicts, r.createErr
}
func (r *fakeTimeOffRepository) ProcessExistingTimeOff(id string, _ bool, _ int64) error {
	r.processed = append(r.processed, id)
	return r.processErrors[id]
}
func (r *fakeTimeOffRepository) GetUnprocessedTimeOffs() ([]timeoffdomain.ExpertTimeOff, error) {
	r.getCalls.Add(1)
	return r.unprocessed, nil
}
func (*fakeTimeOffRepository) GetTimeOffsByExpert(string) ([]timeoffdomain.ExpertTimeOff, error) {
	return nil, nil
}
func (*fakeTimeOffRepository) DeleteTimeOff(string, string) error { return nil }
func (*fakeTimeOffRepository) GetTimeOffs(string, time.Time) ([]timeoffdomain.ExpertTimeOff, error) {
	return nil, nil
}

func TestCreateTimeOffValidatesRangeBeforePersistence(t *testing.T) {
	now := time.UnixMilli(100)
	for _, interval := range [][2]int64{{0, 200}, {200, 0}, {200, 200}, {300, 200}, {1, 100}} {
		repo := &fakeTimeOffRepository{}
		u := newUsecaseWithClock(repo, func() time.Time { return now })
		if _, _, err := u.CreateTimeOff("expert", interval[0], interval[1], "reason"); !errors.Is(err, ErrInvalidDate) {
			t.Fatalf("interval %v: %v", interval, err)
		}
		if repo.createCalls != 0 {
			t.Fatal("invalid interval reached persistence")
		}
	}
}

func TestNormalAndForceCreationUseExplicitRepositoryModes(t *testing.T) {
	now := time.UnixMilli(100)
	for _, force := range []bool{false, true} {
		repo := &fakeTimeOffRepository{}
		u := newUsecaseWithClock(repo, func() time.Time { return now })
		var err error
		if force {
			_, err = u.ConfirmTimeOff("expert", 200, 300, "reason")
		} else {
			_, _, err = u.CreateTimeOff("expert", 200, 300, "reason")
		}
		if err != nil || repo.force != force || repo.created == nil || repo.created.ExpertID != "expert" {
			t.Fatalf("force=%v repo=%+v err=%v", force, repo, err)
		}
	}
}

func TestCreateTimeOffPropagatesTypedConflictAndAffectedIDs(t *testing.T) {
	repo := &fakeTimeOffRepository{createErr: ErrConflict, conflicts: []string{"appt-1"}}
	u := newUsecaseWithClock(repo, func() time.Time { return time.UnixMilli(100) })
	_, conflicts, err := u.CreateTimeOff("expert", 200, 300, "reason")
	if !errors.Is(err, ErrConflict) || len(conflicts) != 1 || conflicts[0] != "appt-1" {
		t.Fatalf("conflicts=%v err=%v", conflicts, err)
	}
}

func TestWorkerProcessesSequentiallyAndContinuesAfterFailure(t *testing.T) {
	repo := &fakeTimeOffRepository{
		unprocessed:   []timeoffdomain.ExpertTimeOff{{TimeOffID: "one"}, {TimeOffID: "two"}},
		processErrors: map[string]error{"one": errors.New("rollback")},
	}
	u := newUsecaseWithClock(repo, time.Now)
	u.ProcessTimeOffs()
	if len(repo.processed) != 2 || repo.processed[0] != "one" || repo.processed[1] != "two" {
		t.Fatalf("worker did not isolate failures: %v", repo.processed)
	}
}

func TestWorkerStopsOnContextCancellation(t *testing.T) {
	repo := &fakeTimeOffRepository{}
	u := newUsecaseWithClock(repo, time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	StartWorkerWithContext(ctx, u, 5*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)
	calls := repo.getCalls.Load()
	time.Sleep(15 * time.Millisecond)
	if after := repo.getCalls.Load(); calls == 0 || after != calls {
		t.Fatalf("worker did not stop: before=%d after=%d", calls, after)
	}
}
