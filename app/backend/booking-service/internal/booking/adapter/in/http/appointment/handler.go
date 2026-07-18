package handler

import (
	appappointment "booking-service/internal/booking/application/appointment"
)

type Handler struct {
	usecase appappointment.Usecase
}

func NewHandler(usecase appappointment.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
