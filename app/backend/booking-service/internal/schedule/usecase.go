package schedule

import (
	"booking-service/internal/booking/domain"
	"booking-service/internal/slot"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSchedule  = errors.New("invalid schedule configuration")
	ErrScheduleOverlap  = errors.New("schedule configuration overlaps")
	ErrScheduleNotFound = errors.New("schedule configuration not found")
)

var businessLocation = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

type Usecase interface {
	CreateTimeTemplate(shiftName, startTime, endTime string, slotDuration int) (*domain.TimeTemplate, error)
	GetTimeTemplates() ([]domain.TimeTemplate, error)
	UpdateTemplate(templateID string, isActive bool) error
	CreateAvailability(expertID, templateID string, dayOfWeek int, effectiveFrom int64, effectiveUntil *int64, price float64) (*domain.Availability, error)
	GetAvailabilities(expertID string) ([]domain.Availability, error)
	UpdateAvailability(availID, expertID string, updates map[string]interface{}) error
}

type TimeOffProvider interface {
	GetTimeOffs(expertID string, fromDate time.Time) ([]domain.ExpertTimeOff, error)
}

type Clock interface{ Now() time.Time }
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type scheduleUsecase struct {
	repo        Repository
	timeOffs    TimeOffProvider
	rollingDays int
	clock       Clock
}

func NewUsecase(repo Repository) Usecase {
	return &scheduleUsecase{repo: repo, clock: realClock{}}
}

func NewUsecaseWithReconciliation(repo Repository, timeOffs TimeOffProvider, rollingDays int) Usecase {
	return &scheduleUsecase{repo: repo, timeOffs: timeOffs, rollingDays: rollingDays, clock: realClock{}}
}

func NewUsecaseWithReconciliationAndClock(repo Repository, timeOffs TimeOffProvider, rollingDays int, clock Clock) Usecase {
	return &scheduleUsecase{repo: repo, timeOffs: timeOffs, rollingDays: rollingDays, clock: clock}
}

func (u *scheduleUsecase) CreateTimeTemplate(shiftName, startTime, endTime string, slotDuration int) (*domain.TimeTemplate, error) {
	template := domain.TimeTemplate{TemplateID: uuid.New().String(), ShiftName: shiftName, StartTime: startTime, EndTime: endTime, SlotDurationMinutes: slotDuration, IsActive: true}
	if err := domain.ValidateTimeTemplate(template); err != nil {
		return nil, mapDomainError(err)
	}
	if err := u.repo.CreateTimeTemplate(&template); err != nil {
		return nil, err
	}
	return &template, nil
}

func (u *scheduleUsecase) GetTimeTemplates() ([]domain.TimeTemplate, error) {
	return u.repo.GetTimeTemplates()
}

func (u *scheduleUsecase) UpdateTemplate(templateID string, isActive bool) error {
	template, err := u.repo.GetTimeTemplateByID(templateID)
	if err != nil {
		return mapRepositoryError(err)
	}
	if err := domain.ValidateTimeTemplate(*template); err != nil {
		return mapDomainError(err)
	}
	if isActive {
		availabilities, err := u.repo.GetEnabledAvailabilitiesByTemplate(templateID)
		if err != nil {
			return err
		}
		for _, availability := range availabilities {
			if _, err := domain.ValidateAvailability(availability, *template); err != nil {
				return mapDomainError(err)
			}
			if err := u.ensureNoAvailabilityOverlap(availability, *template, availability.AvailabilityID); err != nil {
				return err
			}
		}
	}
	updates := map[string]interface{}{"is_active": isActive}
	if u.timeOffs == nil || u.rollingDays == 0 {
		return u.repo.UpdateTemplate(templateID, updates)
	}
	candidateTemplate := *template
	candidateTemplate.IsActive = isActive
	plans, err := u.buildTemplateReconciliationPlans(candidateTemplate)
	if err != nil {
		return err
	}
	_, err = u.repo.ReconcileTemplate(templateID, updates, plans, u.clock.Now().UnixMilli())
	return mapReconciliationError(err)
}

func (u *scheduleUsecase) CreateAvailability(expertID, templateID string, dayOfWeek int, effectiveFrom int64, effectiveUntil *int64, price float64) (*domain.Availability, error) {
	template, err := u.repo.GetTimeTemplateByID(templateID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	availability := domain.Availability{AvailabilityID: uuid.New().String(), ExpertID: expertID, TemplateID: templateID, DayOfWeek: dayOfWeek, IsEnabled: true, EffectiveFrom: effectiveFrom, EffectiveUntil: effectiveUntil, Price: &price}
	if _, err := domain.ValidateAvailability(availability, *template); err != nil {
		return nil, mapDomainError(err)
	}
	if err := u.ensureNoAvailabilityOverlap(availability, *template, ""); err != nil {
		return nil, err
	}
	if err := u.repo.CreateAvailability(&availability); err != nil {
		return nil, err
	}
	return &availability, nil
}

func (u *scheduleUsecase) GetAvailabilities(expertID string) ([]domain.Availability, error) {
	return u.repo.GetAvailabilities(expertID)
}

func (u *scheduleUsecase) UpdateAvailability(availID, expertID string, updates map[string]interface{}) error {
	current, err := u.repo.GetAvailabilityByID(availID, expertID)
	if err != nil {
		return mapRepositoryError(err)
	}
	candidate := *current
	applyAvailabilityUpdates(&candidate, updates)
	template, err := u.repo.GetTimeTemplateByID(candidate.TemplateID)
	if err != nil {
		return mapRepositoryError(err)
	}
	if _, err := domain.ValidateAvailability(candidate, *template); err != nil {
		return mapDomainError(err)
	}
	if candidate.IsEnabled {
		if err := u.ensureNoAvailabilityOverlap(candidate, *template, availID); err != nil {
			return err
		}
	}
	if u.timeOffs == nil || u.rollingDays == 0 {
		return u.repo.UpdateAvailability(availID, expertID, updates)
	}
	candidates, err := u.planAvailability(candidate)
	if err != nil {
		return err
	}
	_, err = u.repo.ReconcileAvailability(candidate, updates, candidates, u.clock.Now().UnixMilli())
	return mapReconciliationError(err)
}

func (u *scheduleUsecase) planAvailability(availability domain.Availability) ([]domain.ExpertSlot, error) {
	templates, err := u.repo.GetAllTimeTemplates()
	if err != nil {
		return nil, err
	}
	timeOffs, err := u.timeOffs.GetTimeOffs(availability.ExpertID, u.clock.Now())
	if err != nil {
		return nil, err
	}
	availabilities := []domain.Availability(nil)
	if availability.IsEnabled {
		availabilities = []domain.Availability{availability}
	}
	candidates, err := slot.PlanSlotsForNextDays(u.clock.Now(), availability.ExpertID, u.rollingDays, availabilities, templates, timeOffs)
	if err != nil {
		return nil, mapReconciliationError(err)
	}
	return candidates, nil
}

func (u *scheduleUsecase) buildTemplateReconciliationPlans(template domain.TimeTemplate) ([]ReconciliationPlan, error) {
	availabilities, err := u.repo.GetEnabledAvailabilitiesByTemplate(template.TemplateID)
	if err != nil {
		return nil, err
	}
	templates, err := u.repo.GetAllTimeTemplates()
	if err != nil {
		return nil, err
	}
	for i := range templates {
		if templates[i].TemplateID == template.TemplateID {
			templates[i] = template
		}
	}
	plans := make([]ReconciliationPlan, 0, len(availabilities))
	for _, availability := range availabilities {
		timeOffs, err := u.timeOffs.GetTimeOffs(availability.ExpertID, u.clock.Now())
		if err != nil {
			return nil, err
		}
		active := []domain.Availability(nil)
		if template.IsActive {
			active = []domain.Availability{availability}
		}
		candidates, err := slot.PlanSlotsForNextDays(u.clock.Now(), availability.ExpertID, u.rollingDays, active, templates, timeOffs)
		if err != nil {
			return nil, mapReconciliationError(err)
		}
		plans = append(plans, ReconciliationPlan{Availability: availability, Candidates: candidates})
	}
	return plans, nil
}

func mapReconciliationError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, slot.ErrSlotOverlap) {
		return fmt.Errorf("%w: %v", ErrScheduleOverlap, err)
	}
	if errors.Is(err, slot.ErrInvalidGeneration) {
		return fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
	}
	return err
}

func (u *scheduleUsecase) ensureNoAvailabilityOverlap(candidate domain.Availability, template domain.TimeTemplate, excludeID string) error {
	existing, err := u.repo.GetEnabledAvailabilities(candidate.ExpertID)
	if err != nil {
		return err
	}
	for _, other := range existing {
		if other.AvailabilityID == excludeID {
			continue
		}
		otherTemplate, err := u.repo.GetTimeTemplateByID(other.TemplateID)
		if err != nil {
			return mapRepositoryError(err)
		}
		overlaps, err := domain.AvailabilitiesOverlap(candidate, template, other, *otherTemplate, businessLocation)
		if err != nil {
			return mapDomainError(err)
		}
		if overlaps {
			return fmt.Errorf("%w: expert weekday schedules overlap", ErrScheduleOverlap)
		}
	}
	return nil
}

func applyAvailabilityUpdates(availability *domain.Availability, updates map[string]interface{}) {
	if value, ok := updates["template_id"].(string); ok {
		availability.TemplateID = value
	}
	if value, ok := updates["day_of_week"].(int); ok {
		availability.DayOfWeek = value
	}
	if value, ok := updates["is_enabled"].(bool); ok {
		availability.IsEnabled = value
	}
	if value, ok := updates["effective_from"].(int64); ok {
		availability.EffectiveFrom = value
	}
	if value, ok := updates["effective_until"].(int64); ok {
		availability.EffectiveUntil = &value
	}
	if value, ok := updates["price"].(float64); ok {
		availability.Price = &value
	}
}

func mapDomainError(err error) error {
	if errors.Is(err, domain.ErrScheduleOverlap) {
		return fmt.Errorf("%w: %v", ErrScheduleOverlap, err)
	}
	if errors.Is(err, domain.ErrInvalidSchedule) || errors.Is(err, domain.ErrInvalidMoneyVND) {
		return fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
	}
	return err
}

func mapRepositoryError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return ErrScheduleNotFound
	}
	return err
}
