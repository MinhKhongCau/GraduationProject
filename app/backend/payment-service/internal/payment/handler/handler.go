package handler

import (
	"payment-service/internal/payment"
)

type Handler struct {
	usecase payment.Usecase
}

func NewHandler(usecase payment.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
