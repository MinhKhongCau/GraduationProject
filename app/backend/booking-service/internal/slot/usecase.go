package slot

import (
	"booking-service/internal/booking/domain"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

const GenerationLeadTime = 5 * time.Minute

var generationLocation = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

type Clock interface{ Now() time.Time }
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type GenerationResult struct {
	Candidates int
	Inserted   int64
}

type Usecase interface {
	GenerateSlotsForNextDays(expertID string, daysToGenerate int, availabilities []domain.Availability, timeTemplates []domain.TimeTemplate, timeOffs []domain.ExpertTimeOff) (GenerationResult, error)
	LockSlot(slotID, patientID string) error
	GetDates(expertID string, startDate, endDate time.Time) ([]string, error)
	GetTimes(expertID string, date string) ([]SlotTimeResult, error)
}

type slotUsecase struct {
	repo            Repository
	appointmentRepo appointmentRepository
	clock           Clock
	generationMu    sync.Mutex
}

type appointmentRepository interface {
	LockSlot(slotID, patientID string) error
}

func NewUsecase(repo Repository, appointmentRepo appointmentRepository) Usecase {
	return NewUsecaseWithClock(repo, appointmentRepo, realClock{})
}

func NewUsecaseWithClock(repo Repository, appointmentRepo appointmentRepository, clock Clock) Usecase {
	return &slotUsecase{repo: repo, appointmentRepo: appointmentRepo, clock: clock}
}

func (u *slotUsecase) LockSlot(slotID, patientID string) error {
	if err := u.appointmentRepo.LockSlot(slotID, patientID); err != nil {
		return ErrSlotAlreadyLocked
	}
	return nil
}

func (u *slotUsecase) GetDates(expertID string, startDate, endDate time.Time) ([]string, error) {
	return u.repo.GetAvailableDates(expertID, startDate, endDate)
}

func (u *slotUsecase) GetTimes(expertID, date string) ([]SlotTimeResult, error) {
	return u.repo.GetAvailableTimes(date, expertID)
}

func (u *slotUsecase) GenerateSlotsForNextDays(expertID string, daysToGenerate int, availabilities []domain.Availability, templates []domain.TimeTemplate, timeOffs []domain.ExpertTimeOff) (GenerationResult, error) {
	if daysToGenerate <= 0 {
		return GenerationResult{}, fmt.Errorf("%w: days_to_generate must be positive", ErrInvalidGeneration)
	}
	u.generationMu.Lock()
	defer u.generationMu.Unlock()

	candidates, err := PlanSlotsForNextDays(u.clock.Now(), expertID, daysToGenerate, availabilities, templates, timeOffs)
	if err != nil {
		return GenerationResult{}, err
	}
	result := GenerationResult{Candidates: len(candidates)}
	if len(candidates) == 0 {
		return result, nil
	}

	existing, err := u.repo.GetOverlappingSlots(expertID, candidates[0].StartTime, maxEnd(candidates))
	if err != nil {
		return GenerationResult{}, err
	}
	toInsert, err := removeExactDuplicatesAndRejectOverlaps(candidates, existing)
	if err != nil {
		return GenerationResult{}, err
	}
	inserted, err := u.repo.BulkInsertSlots(toInsert)
	if err != nil {
		return GenerationResult{}, err
	}
	result.Inserted = inserted
	return result, nil
}

// PlanSlotsForNextDays deterministically builds valid slot candidates without persistence.
// Reconciliation uses it before committing configuration and slot changes atomically.
func PlanSlotsForNextDays(now time.Time, expertID string, daysToGenerate int, availabilities []domain.Availability, templates []domain.TimeTemplate, timeOffs []domain.ExpertTimeOff) ([]domain.ExpertSlot, error) {
	if daysToGenerate <= 0 {
		return nil, fmt.Errorf("%w: days_to_generate must be positive", ErrInvalidGeneration)
	}
	now = now.In(generationLocation)
	cutoff := now.Add(GenerationLeadTime)
	templateByID := make(map[string]domain.TimeTemplate, len(templates))
	for _, template := range templates {
		templateByID[template.TemplateID] = template
	}

	var candidates []domain.ExpertSlot
	for dayOffset := 0; dayOffset < daysToGenerate; dayOffset++ {
		targetDate := now.AddDate(0, 0, dayOffset)
		weekday := dayOfWeek(targetDate)
		for _, availability := range availabilities {
			if availability.DayOfWeek != weekday || !domain.AvailabilityAppliesOnDate(availability, targetDate, generationLocation) {
				continue
			}
			template, ok := templateByID[availability.TemplateID]
			if !ok {
				return nil, fmt.Errorf("%w: template %s not found", ErrInvalidGeneration, availability.TemplateID)
			}
			if !template.IsActive {
				continue
			}
			price, err := domain.ValidateAvailability(availability, template)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrInvalidGeneration, err)
			}
			slots, err := sliceShiftIntoSlots(expertID, availability.AvailabilityID, targetDate, template, float64(price), cutoff)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrInvalidGeneration, err)
			}
			for _, candidate := range slots {
				if !overlapsAnyTimeOff(candidate, timeOffs) {
					candidates = append(candidates, candidate)
				}
			}
		}
	}

	sortSlots(candidates)
	if err := rejectCandidateOverlaps(candidates); err != nil {
		return nil, err
	}
	return candidates, nil
}

func sliceShiftIntoSlots(expertID, availabilityID string, targetDate time.Time, template domain.TimeTemplate, price float64, cutoff time.Time) ([]domain.ExpertSlot, error) {
	if err := domain.ValidateTimeTemplate(template); err != nil {
		return nil, err
	}
	if _, err := domain.NewMoneyVNDFromPrice(price); err != nil {
		return nil, err
	}
	startClock, _ := domain.ParseClockTime(template.StartTime)
	endClock, _ := domain.ParseClockTime(template.EndTime)
	localDate := targetDate.In(generationLocation)
	shiftStart := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), startClock.Hour, startClock.Minute, 0, 0, generationLocation)
	shiftEnd := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), endClock.Hour, endClock.Minute, 0, 0, generationLocation)
	nowMs := cutoff.Add(-GenerationLeadTime).UnixMilli()
	dateOnly := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), 0, 0, 0, 0, generationLocation)
	var slots []domain.ExpertSlot
	for start := shiftStart; start.Before(shiftEnd); start = start.Add(time.Duration(template.SlotDurationMinutes) * time.Minute) {
		end := start.Add(time.Duration(template.SlotDurationMinutes) * time.Minute)
		if end.After(shiftEnd) {
			break
		}
		if !start.After(cutoff) {
			continue
		}
		availability := availabilityID
		slots = append(slots, domain.ExpertSlot{SlotID: uuid.New().String(), ExpertID: expertID, AvailabilityID: &availability, DateSlot: dateOnly, StartTime: start.UnixMilli(), EndTime: end.UnixMilli(), Status: domain.SlotStatusAvailable, Price: price, CreatedAt: nowMs, UpdatedAt: nowMs})
	}
	return slots, nil
}

func overlapsAnyTimeOff(slot domain.ExpertSlot, timeOffs []domain.ExpertTimeOff) bool {
	for _, timeOff := range timeOffs {
		if domain.IntervalsOverlap(slot.StartTime, slot.EndTime, timeOff.StartDatetime, timeOff.EndDatetime) {
			return true
		}
	}
	return false
}

func rejectCandidateOverlaps(slots []domain.ExpertSlot) error {
	for i := 1; i < len(slots); i++ {
		if domain.IntervalsOverlap(slots[i-1].StartTime, slots[i-1].EndTime, slots[i].StartTime, slots[i].EndTime) {
			return fmt.Errorf("%w: candidates %d-%d and %d-%d", ErrSlotOverlap, slots[i-1].StartTime, slots[i-1].EndTime, slots[i].StartTime, slots[i].EndTime)
		}
	}
	return nil
}

func removeExactDuplicatesAndRejectOverlaps(candidates, existing []domain.ExpertSlot) ([]domain.ExpertSlot, error) {
	filtered := make([]domain.ExpertSlot, 0, len(candidates))
	for _, candidate := range candidates {
		exact := false
		for _, persisted := range existing {
			if candidate.StartTime == persisted.StartTime && candidate.EndTime == persisted.EndTime {
				exact = true
				break
			}
			if domain.IntervalsOverlap(candidate.StartTime, candidate.EndTime, persisted.StartTime, persisted.EndTime) {
				return nil, fmt.Errorf("%w: candidate %d-%d conflicts with existing slot %s", ErrSlotOverlap, candidate.StartTime, candidate.EndTime, persisted.SlotID)
			}
		}
		if !exact {
			filtered = append(filtered, candidate)
		}
	}
	return filtered, nil
}

func sortSlots(slots []domain.ExpertSlot) {
	sort.Slice(slots, func(i, j int) bool {
		if slots[i].StartTime == slots[j].StartTime {
			return slots[i].EndTime < slots[j].EndTime
		}
		return slots[i].StartTime < slots[j].StartTime
	})
}
func maxEnd(slots []domain.ExpertSlot) int64 {
	maximum := slots[0].EndTime
	for _, slot := range slots[1:] {
		if slot.EndTime > maximum {
			maximum = slot.EndTime
		}
	}
	return maximum
}
func dayOfWeek(date time.Time) int {
	day := int(date.Weekday())
	if day == 0 {
		return 7
	}
	return day
}
