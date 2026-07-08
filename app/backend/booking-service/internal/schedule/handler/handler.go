package handler

import (
	"booking-service/internal/schedule"
)

type Handler struct {
	usecase schedule.Usecase
}

func NewHandler(usecase schedule.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
