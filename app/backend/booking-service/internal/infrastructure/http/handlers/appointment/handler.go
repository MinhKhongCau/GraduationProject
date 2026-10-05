package handler

import (
	appappointment "booking-service/internal/application/appointment"
)

type Handler struct {
	usecase appappointment.Usecase
	reader  appappointment.ReadUsecase
}

func NewHandler(usecase appappointment.Usecase) *Handler {
	handler := &Handler{usecase: usecase}
	if reader, ok := usecase.(appappointment.ReadUsecase); ok {
		handler.reader = reader
	}
	return handler
}
