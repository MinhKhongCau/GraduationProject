package handler

import (
	"booking-service/internal/application/schedule"
)

type Handler struct {
	usecase schedule.Usecase
}

func NewHandler(usecase schedule.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
