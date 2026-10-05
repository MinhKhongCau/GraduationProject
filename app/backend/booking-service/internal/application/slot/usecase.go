package slot

import (
	bookingquery "booking-service/internal/application/query"
	scheduledomain "booking-service/internal/domain/schedule"
	"booking-service/internal/domain/shared"
	slotdomain "booking-service/internal/domain/slot"
	timeoffdomain "booking-service/internal/domain/timeoff"
	"errors"
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
	GenerateSlotsForNextDays(expertID string, daysToGenerate int, availabilities []scheduledomain.Availability, timeTemplates []scheduledomain.TimeTemplate, timeOffs []timeoffdomain.ExpertTimeOff) (GenerationResult, error)
	LockSlot(slotID, patientID string) error
	GetDates(expertID string, startDate, endDate time.Time) ([]string, error)
	GetTimes(expertID string, date string) ([]SlotTimeResult, error)
}

type ReadUsecase interface {
	ListAvailableDates(query AvailableDateQuery) (bookingquery.Page[string], error)
	ListAvailableTimes(query AvailableTimeQuery) (bookingquery.Page[SlotTimeResult], error)
	ListExpertSlots(query ExpertSlotQuery) (bookingquery.Page[slotdomain.ExpertSlot], error)
}

type readRepository interface {
	ListAvailableDates(query AvailableDateQuery) ([]string, int64, error)
	ListAvailableTimes(query AvailableTimeQuery) ([]SlotTimeResult, int64, error)
	ListExpertSlots(query ExpertSlotQuery) ([]slotdomain.ExpertSlot, int64, error)
}

type AvailableDateQuery struct {
	ExpertID string
	FromMs   int64
	ToMs     int64
	Page     bookingquery.PageRequest
}
type AvailableTimeQuery struct {
	ExpertID string
	Date     string
	Page     bookingquery.PageRequest
}
type ExpertSlotQuery struct {
	ExpertID       string
	FromMs         int64
	ToMs           int64
	Status         *slotdomain.SlotStatus
	AvailabilityID string
	Page           bookingquery.PageRequest
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

func (u *slotUsecase) ListAvailableDates(filter AvailableDateQuery) (bookingquery.Page[string], error) {
	reader, ok := u.repo.(readRepository)
	if !ok {
		return bookingquery.Page[string]{}, errors.New("slot reader unavailable")
	}
	items, total, err := reader.ListAvailableDates(filter)
	if err != nil {
		return bookingquery.Page[string]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *slotUsecase) ListAvailableTimes(filter AvailableTimeQuery) (bookingquery.Page[SlotTimeResult], error) {
	reader, ok := u.repo.(readRepository)
	if !ok {
		return bookingquery.Page[SlotTimeResult]{}, errors.New("slot reader unavailable")
	}
	items, total, err := reader.ListAvailableTimes(filter)
	if err != nil {
		return bookingquery.Page[SlotTimeResult]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *slotUsecase) ListExpertSlots(filter ExpertSlotQuery) (bookingquery.Page[slotdomain.ExpertSlot], error) {
	reader, ok := u.repo.(readRepository)
	if !ok {
		return bookingquery.Page[slotdomain.ExpertSlot]{}, errors.New("slot reader unavailable")
	}
	items, total, err := reader.ListExpertSlots(filter)
	if err != nil {
		return bookingquery.Page[slotdomain.ExpertSlot]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *slotUsecase) GenerateSlotsForNextDays(expertID string, daysToGenerate int, availabilities []scheduledomain.Availability, templates []scheduledomain.TimeTemplate, timeOffs []timeoffdomain.ExpertTimeOff) (GenerationResult, error) {
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
func PlanSlotsForNextDays(now time.Time, expertID string, daysToGenerate int, availabilities []scheduledomain.Availability, templates []scheduledomain.TimeTemplate, timeOffs []timeoffdomain.ExpertTimeOff) ([]slotdomain.ExpertSlot, error) {
	if daysToGenerate <= 0 {
		return nil, fmt.Errorf("%w: days_to_generate must be positive", ErrInvalidGeneration)
	}
	now = now.In(generationLocation)
	cutoff := now.Add(GenerationLeadTime)
	templateByID := make(map[string]scheduledomain.TimeTemplate, len(templates))
	for _, template := range templates {
		templateByID[template.TemplateID] = template
	}

	var candidates []slotdomain.ExpertSlot
	for dayOffset := 0; dayOffset < daysToGenerate; dayOffset++ {
		targetDate := now.AddDate(0, 0, dayOffset)
		weekday := dayOfWeek(targetDate)
		for _, availability := range availabilities {
			if availability.DayOfWeek != weekday || !scheduledomain.AvailabilityAppliesOnDate(availability, targetDate, generationLocation) {
				continue
			}
			template, ok := templateByID[availability.TemplateID]
			if !ok {
				return nil, fmt.Errorf("%w: template %s not found", ErrInvalidGeneration, availability.TemplateID)
			}
			if !template.IsActive {
				continue
			}
			price, err := scheduledomain.ValidateAvailability(availability, template)
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

func sliceShiftIntoSlots(expertID, availabilityID string, targetDate time.Time, template scheduledomain.TimeTemplate, price float64, cutoff time.Time) ([]slotdomain.ExpertSlot, error) {
	if err := scheduledomain.ValidateTimeTemplate(template); err != nil {
		return nil, err
	}
	if _, err := shared.NewMoneyVNDFromPrice(price); err != nil {
		return nil, err
	}
	startClock, _ := scheduledomain.ParseClockTime(template.StartTime)
	endClock, _ := scheduledomain.ParseClockTime(template.EndTime)
	localDate := targetDate.In(generationLocation)
	shiftStart := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), startClock.Hour, startClock.Minute, 0, 0, generationLocation)
	shiftEnd := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), endClock.Hour, endClock.Minute, 0, 0, generationLocation)
	nowMs := cutoff.Add(-GenerationLeadTime).UnixMilli()
	dateOnly := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), 0, 0, 0, 0, generationLocation)
	var slots []slotdomain.ExpertSlot
	for start := shiftStart; start.Before(shiftEnd); start = start.Add(time.Duration(template.SlotDurationMinutes) * time.Minute) {
		end := start.Add(time.Duration(template.SlotDurationMinutes) * time.Minute)
		if end.After(shiftEnd) {
			break
		}
		if !start.After(cutoff) {
			continue
		}
		availability := availabilityID
		slots = append(slots, slotdomain.ExpertSlot{SlotID: uuid.New().String(), ExpertID: expertID, AvailabilityID: &availability, DateSlot: dateOnly, StartTime: start.UnixMilli(), EndTime: end.UnixMilli(), Status: slotdomain.SlotStatusAvailable, Price: price, CreatedAt: nowMs, UpdatedAt: nowMs})
	}
	return slots, nil
}

func overlapsAnyTimeOff(slot slotdomain.ExpertSlot, timeOffs []timeoffdomain.ExpertTimeOff) bool {
	for _, timeOff := range timeOffs {
		if shared.IntervalsOverlap(slot.StartTime, slot.EndTime, timeOff.StartDatetime, timeOff.EndDatetime) {
			return true
		}
	}
	return false
}

func rejectCandidateOverlaps(slots []slotdomain.ExpertSlot) error {
	for i := 1; i < len(slots); i++ {
		if shared.IntervalsOverlap(slots[i-1].StartTime, slots[i-1].EndTime, slots[i].StartTime, slots[i].EndTime) {
			return fmt.Errorf("%w: candidates %d-%d and %d-%d", ErrSlotOverlap, slots[i-1].StartTime, slots[i-1].EndTime, slots[i].StartTime, slots[i].EndTime)
		}
	}
	return nil
}

func removeExactDuplicatesAndRejectOverlaps(candidates, existing []slotdomain.ExpertSlot) ([]slotdomain.ExpertSlot, error) {
	filtered := make([]slotdomain.ExpertSlot, 0, len(candidates))
	for _, candidate := range candidates {
		exact := false
		for _, persisted := range existing {
			if candidate.StartTime == persisted.StartTime && candidate.EndTime == persisted.EndTime {
				exact = true
				break
			}
			if shared.IntervalsOverlap(candidate.StartTime, candidate.EndTime, persisted.StartTime, persisted.EndTime) {
				return nil, fmt.Errorf("%w: candidate %d-%d conflicts with existing slot %s", ErrSlotOverlap, candidate.StartTime, candidate.EndTime, persisted.SlotID)
			}
		}
		if !exact {
			filtered = append(filtered, candidate)
		}
	}
	return filtered, nil
}

func sortSlots(slots []slotdomain.ExpertSlot) {
	sort.Slice(slots, func(i, j int) bool {
		if slots[i].StartTime == slots[j].StartTime {
			return slots[i].EndTime < slots[j].EndTime
		}
		return slots[i].StartTime < slots[j].StartTime
	})
}
func maxEnd(slots []slotdomain.ExpertSlot) int64 {
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
