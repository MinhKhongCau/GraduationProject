package handler

import (
	apppayment "payment-service/internal/payment/application"
)

type Handler struct {
	usecase apppayment.Usecase
}

func NewHandler(usecase apppayment.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
