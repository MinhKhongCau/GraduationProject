package schedule

import (
	scheduledomain "booking-service/internal/domain/schedule"
	slotdomain "booking-service/internal/domain/slot"
)

type ReconciliationPlan struct {
	Availability scheduledomain.Availability
	Candidates   []slotdomain.ExpertSlot
}

type Repository interface {
	GetAvailabilities(expertID string) ([]scheduledomain.Availability, error)
	GetTimeTemplates() ([]scheduledomain.TimeTemplate, error)
	GetAllTimeTemplates() ([]scheduledomain.TimeTemplate, error)
	GetTimeTemplateByID(templateID string) (*scheduledomain.TimeTemplate, error)
	GetAvailabilityByID(availID, expertID string) (*scheduledomain.Availability, error)
	GetEnabledAvailabilities(expertID string) ([]scheduledomain.Availability, error)
	GetEnabledAvailabilitiesByTemplate(templateID string) ([]scheduledomain.Availability, error)
	GetEnabledExpertIDs() ([]string, error)
	CreateTimeTemplate(template *scheduledomain.TimeTemplate) error
	CreateAvailability(avail *scheduledomain.Availability) error
	UpdateAvailability(availID string, expertID string, updates map[string]interface{}) error
	UpdateTemplate(templateID string, updates map[string]interface{}) error
	ReconcileAvailability(availability scheduledomain.Availability, updates map[string]interface{}, candidates []slotdomain.ExpertSlot, nowMs int64) (int64, error)
	ReconcileTemplate(templateID string, updates map[string]interface{}, plans []ReconciliationPlan, nowMs int64) (int64, error)
}
