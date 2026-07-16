package handler

import (
	"payment-service/internal/wallet"
)

type Handler struct {
	usecase wallet.Usecase
}

func NewHandler(usecase wallet.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
