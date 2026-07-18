package handler

import (
	"booking-service/internal/appointment"
)

type Handler struct {
	usecase appointment.Usecase
}

func NewHandler(usecase appointment.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
