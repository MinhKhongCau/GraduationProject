package slot

import (
	"booking-service/internal/booking/domain"
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeGenerationProvider struct {
	experts []string
	avails  map[string][]domain.Availability
	err     error
}

func (p *fakeGenerationProvider) GetAvailabilities(expertID string) ([]domain.Availability, error) {
	return p.avails[expertID], nil
}
func (*fakeGenerationProvider) GetAllTimeTemplates() ([]domain.TimeTemplate, error) {
	return []domain.TimeTemplate{{TemplateID: "tpl", IsActive: true}}, nil
}
func (p *fakeGenerationProvider) GetEnabledExpertIDs() ([]string, error) { return p.experts, p.err }

type fakeGenerationTimeOffs struct{}

func (fakeGenerationTimeOffs) GetTimeOffs(string, time.Time) ([]domain.ExpertTimeOff, error) {
	return nil, nil
}

type recordingGenerator struct {
	mu       sync.Mutex
	calls    []string
	fail     map[string]error
	inserted map[string]int64
}

func (g *recordingGenerator) GenerateSlotsForNextDays(expertID string, _ int, _ []domain.Availability, _ []domain.TimeTemplate, _ []domain.ExpertTimeOff) (GenerationResult, error) {
	g.mu.Lock()
	g.calls = append(g.calls, expertID)
	g.mu.Unlock()
	if err := g.fail[expertID]; err != nil {
		return GenerationResult{}, err
	}
	return GenerationResult{Candidates: 1, Inserted: g.inserted[expertID]}, nil
}
func (*recordingGenerator) LockSlot(string, string) error { return nil }
func (*recordingGenerator) GetDates(string, time.Time, time.Time) ([]string, error) {
	return nil, nil
}
func (*recordingGenerator) GetTimes(string, string) ([]SlotTimeResult, error) { return nil, nil }

func TestGenerationServiceIsolatesExpertsAndSummarizesRun(t *testing.T) {
	generator := &recordingGenerator{fail: map[string]error{"bad": errors.New("invalid price")}, inserted: map[string]int64{"good": 4}}
	provider := &fakeGenerationProvider{experts: []string{"bad", "good"}, avails: map[string][]domain.Availability{}}
	service := NewGenerationService(generator, provider, fakeGenerationTimeOffs{})

	summary := service.GenerateAll(context.Background(), 30)
	if summary.Processed != 2 || summary.Succeeded != 1 || summary.Failed != 1 || summary.Inserted != 4 || len(summary.Errors) != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(generator.calls) != 2 || generator.calls[1] != "good" {
		t.Fatalf("later expert was not processed: %v", generator.calls)
	}
}

type workerGenerationApp struct {
	mu    sync.Mutex
	calls int
	seen  chan struct{}
}

func (*workerGenerationApp) GenerateExpert(context.Context, string, int) (GenerationResult, error) {
	return GenerationResult{}, nil
}
func (a *workerGenerationApp) GenerateAll(context.Context, int) GenerationRunSummary {
	a.mu.Lock()
	a.calls++
	a.mu.Unlock()
	select {
	case a.seen <- struct{}{}:
	default:
	}
	return GenerationRunSummary{Processed: 1, Succeeded: 1}
}

func TestStartupAndDailyGenerationAndCancellation(t *testing.T) {
	app := &workerGenerationApp{seen: make(chan struct{}, 4)}
	RunStartupGeneration(context.Background(), app, 30)
	ctx, cancel := context.WithCancel(context.Background())
	StartRollingGenerationWorkerWithInterval(ctx, app, 30, 5*time.Millisecond)
	select {
	case <-app.seen:
	case <-time.After(time.Second):
		t.Fatal("startup invocation not observed")
	}
	select {
	case <-app.seen:
	case <-time.After(time.Second):
		t.Fatal("daily invocation not observed")
	}
	cancel()
	time.Sleep(20 * time.Millisecond)
	app.mu.Lock()
	callsAfterCancel := app.calls
	app.mu.Unlock()
	time.Sleep(20 * time.Millisecond)
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.calls != callsAfterCancel || app.calls < 2 {
		t.Fatalf("worker did not stop cleanly: before=%d after=%d", callsAfterCancel, app.calls)
	}
}
