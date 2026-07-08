package handler

import (
	"booking-service/internal/appointment"
	"booking-service/internal/domain"
	"booking-service/internal/slot"
	"time"
)

type ScheduleProvider interface {
	GetAvailabilities(expertID string) ([]domain.Availability, error)
	GetTimeTemplates() ([]domain.TimeTemplate, error)
}

type TimeOffProvider interface {
	GetTimeOffs(expertID string, fromDate time.Time) ([]domain.ExpertTimeOff, error)
}

type Handler struct {
	repo            slot.Repository
	appointmentRepo appointment.Repository
	usecase         slot.Usecase
	scheduleRepo    ScheduleProvider
	timeoffRepo     TimeOffProvider
}

func NewHandler(
	repo slot.Repository,
	appointmentRepo appointment.Repository,
	usecase slot.Usecase,
	scheduleRepo ScheduleProvider,
	timeoffRepo TimeOffProvider,
) *Handler {
	return &Handler{
		repo:            repo,
		appointmentRepo: appointmentRepo,
		usecase:         usecase,
		scheduleRepo:    scheduleRepo,
		timeoffRepo:     timeoffRepo,
	}
}
