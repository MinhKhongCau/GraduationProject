package handler

import (
	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/application/slot"
)

type Handler struct {
	repo            slot.Repository
	appointmentRepo appappointment.Repository
	usecase         slot.Usecase
	reader          slot.ReadUsecase
	generation      slot.GenerationApplication
}

func NewHandler(
	repo slot.Repository,
	appointmentRepo appappointment.Repository,
	usecase slot.Usecase,
	generation slot.GenerationApplication,
) *Handler {
	handler := &Handler{
		repo:            repo,
		appointmentRepo: appointmentRepo,
		usecase:         usecase,
		generation:      generation,
	}
	if reader, ok := usecase.(slot.ReadUsecase); ok {
		handler.reader = reader
	}
	return handler
}
