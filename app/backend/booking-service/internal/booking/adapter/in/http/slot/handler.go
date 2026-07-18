package handler

import (
	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/booking/domain"
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
	appointmentRepo appappointment.Repository
	usecase         slot.Usecase
	scheduleRepo    ScheduleProvider
	timeoffRepo     TimeOffProvider
}

func NewHandler(
	repo slot.Repository,
	appointmentRepo appappointment.Repository,
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
