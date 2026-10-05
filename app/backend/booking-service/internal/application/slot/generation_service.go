package slot

import (
	scheduledomain "booking-service/internal/domain/schedule"
	timeoffdomain "booking-service/internal/domain/timeoff"
	"context"
	"fmt"
	"time"
)

type GenerationScheduleProvider interface {
	GetAvailabilities(expertID string) ([]scheduledomain.Availability, error)
	GetAllTimeTemplates() ([]scheduledomain.TimeTemplate, error)
	GetEnabledExpertIDs() ([]string, error)
}

type GenerationTimeOffProvider interface {
	GetTimeOffs(expertID string, fromDate time.Time) ([]timeoffdomain.ExpertTimeOff, error)
}

type GenerationApplication interface {
	GenerateExpert(ctx context.Context, expertID string, days int) (GenerationResult, error)
	GenerateAll(ctx context.Context, days int) GenerationRunSummary
}

type ExpertGenerationError struct {
	ExpertID string
	Err      error
}

type GenerationRunSummary struct {
	Processed int
	Succeeded int
	Failed    int
	Inserted  int64
	Errors    []ExpertGenerationError
}

type generationService struct {
	generator Usecase
	schedules GenerationScheduleProvider
	timeOffs  GenerationTimeOffProvider
	clock     Clock
}

func NewGenerationService(generator Usecase, schedules GenerationScheduleProvider, timeOffs GenerationTimeOffProvider) GenerationApplication {
	return &generationService{generator: generator, schedules: schedules, timeOffs: timeOffs, clock: realClock{}}
}

func NewGenerationServiceWithClock(generator Usecase, schedules GenerationScheduleProvider, timeOffs GenerationTimeOffProvider, clock Clock) GenerationApplication {
	return &generationService{generator: generator, schedules: schedules, timeOffs: timeOffs, clock: clock}
}

func (s *generationService) GenerateExpert(ctx context.Context, expertID string, days int) (GenerationResult, error) {
	if err := ctx.Err(); err != nil {
		return GenerationResult{}, err
	}
	availabilities, err := s.schedules.GetAvailabilities(expertID)
	if err != nil {
		return GenerationResult{}, fmt.Errorf("load expert availabilities: %w", err)
	}
	templates, err := s.schedules.GetAllTimeTemplates()
	if err != nil {
		return GenerationResult{}, fmt.Errorf("load time templates: %w", err)
	}
	timeOffs, err := s.timeOffs.GetTimeOffs(expertID, s.clock.Now())
	if err != nil {
		return GenerationResult{}, fmt.Errorf("load expert time-offs: %w", err)
	}
	return s.generator.GenerateSlotsForNextDays(expertID, days, availabilities, templates, timeOffs)
}

func (s *generationService) GenerateAll(ctx context.Context, days int) GenerationRunSummary {
	summary := GenerationRunSummary{}
	expertIDs, err := s.schedules.GetEnabledExpertIDs()
	if err != nil {
		summary.Failed = 1
		summary.Errors = append(summary.Errors, ExpertGenerationError{Err: fmt.Errorf("load enabled experts: %w", err)})
		return summary
	}
	for _, expertID := range expertIDs {
		if ctx.Err() != nil {
			break
		}
		summary.Processed++
		result, err := s.GenerateExpert(ctx, expertID, days)
		if err != nil {
			summary.Failed++
			summary.Errors = append(summary.Errors, ExpertGenerationError{ExpertID: expertID, Err: err})
			continue
		}
		summary.Succeeded++
		summary.Inserted += result.Inserted
	}
	return summary
}
