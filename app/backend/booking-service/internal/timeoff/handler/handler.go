package handler

import (
	"booking-service/internal/timeoff"
)

type Handler struct {
	usecase timeoff.Usecase
}

func NewHandler(usecase timeoff.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}
