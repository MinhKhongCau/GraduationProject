package handler

import (
	apppayment "payment-service/internal/payment/application"
)

type Handler struct {
	usecase apppayment.Usecase
	reader  apppayment.ReadUsecase
}

func NewHandler(usecase apppayment.Usecase) *Handler {
	handler := &Handler{usecase: usecase}
	if reader, ok := usecase.(apppayment.ReadUsecase); ok {
		handler.reader = reader
	}
	return handler
}
