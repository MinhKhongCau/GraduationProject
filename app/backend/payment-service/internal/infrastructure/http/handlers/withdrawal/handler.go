package withdrawal

import (
	"payment-service/internal/application/withdrawal"
)

type Handler struct {
	usecase withdrawal.Usecase
}

func NewHandler(usecase withdrawal.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
