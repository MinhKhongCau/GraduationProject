package handler

import (
	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/slot"
)

type Handler struct {
	repo            slot.Repository
	appointmentRepo appappointment.Repository
	usecase         slot.Usecase
	generation      slot.GenerationApplication
}

func NewHandler(
	repo slot.Repository,
	appointmentRepo appappointment.Repository,
	usecase slot.Usecase,
	generation slot.GenerationApplication,
) *Handler {
	return &Handler{
		repo:            repo,
		appointmentRepo: appointmentRepo,
		usecase:         usecase,
		generation:      generation,
	}
}
