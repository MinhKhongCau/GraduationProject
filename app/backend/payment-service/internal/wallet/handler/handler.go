package handler

import (
	"payment-service/internal/wallet"
)

type Handler struct {
	usecase wallet.Usecase
	reader  wallet.ReadUsecase
}

func NewHandler(usecase wallet.Usecase) *Handler {
	handler := &Handler{usecase: usecase}
	if reader, ok := usecase.(wallet.ReadUsecase); ok {
		handler.reader = reader
	}
	return handler
}
