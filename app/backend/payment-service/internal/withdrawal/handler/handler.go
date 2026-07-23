package handler

import (
	"payment-service/internal/withdrawal"
)

type Handler struct {
	usecase withdrawal.Usecase
}

func NewHandler(usecase withdrawal.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
