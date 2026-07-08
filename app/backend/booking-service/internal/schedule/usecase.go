package schedule

import (
	"booking-service/internal/domain"
	"github.com/google/uuid"
)

type Usecase interface {
	CreateTimeTemplate(shiftName, startTime, endTime string, slotDuration int) (*domain.TimeTemplate, error)
	GetTimeTemplates() ([]domain.TimeTemplate, error)
	UpdateTemplate(templateID string, isActive bool) error

	CreateAvailability(expertID, templateID string, dayOfWeek int, effectiveFrom int64, effectiveUntil *int64) (*domain.Availability, error)
	GetAvailabilities(expertID string) ([]domain.Availability, error)
	UpdateAvailability(availID, expertID string, updates map[string]interface{}) error
}

type scheduleUsecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &scheduleUsecase{repo: repo}
}

func (u *scheduleUsecase) CreateTimeTemplate(shiftName, startTime, endTime string, slotDuration int) (*domain.TimeTemplate, error) {
	template := domain.TimeTemplate{
		TemplateID:          uuid.New().String(),
		ShiftName:           shiftName,
		StartTime:           startTime,
		EndTime:             endTime,
		SlotDurationMinutes: slotDuration,
		IsActive:            true,
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
	updates := map[string]interface{}{
		"is_active": isActive,
	}
	// TODO: check if template exists before updating to return ErrTemplateNotFound
	return u.repo.UpdateTemplate(templateID, updates)
}

func (u *scheduleUsecase) CreateAvailability(expertID, templateID string, dayOfWeek int, effectiveFrom int64, effectiveUntil *int64) (*domain.Availability, error) {
	avail := domain.Availability{
		AvailabilityID: uuid.New().String(),
		ExpertID:       expertID,
		TemplateID:     templateID,
		DayOfWeek:      dayOfWeek,
		IsEnabled:      true,
		EffectiveFrom:  effectiveFrom,
		EffectiveUntil: effectiveUntil,
	}

	if err := u.repo.CreateAvailability(&avail); err != nil {
		return nil, err
	}
	return &avail, nil
}

func (u *scheduleUsecase) GetAvailabilities(expertID string) ([]domain.Availability, error) {
	return u.repo.GetAvailabilities(expertID)
}

func (u *scheduleUsecase) UpdateAvailability(availID, expertID string, updates map[string]interface{}) error {
	// TODO: check if availability exists
	return u.repo.UpdateAvailability(availID, expertID, updates)
}
