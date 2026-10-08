package wallet

import (
	"payment-service/internal/application/wallet"
)

type Handler struct {
	usecase wallet.Usecase
	reader  wallet.ReadUsecase
	managed *wallet.ManagedReader
}

func NewHandler(usecase wallet.Usecase) *Handler {
	handler := &Handler{usecase: usecase}
	if reader, ok := usecase.(wallet.ReadUsecase); ok {
		handler.reader = reader
	}
	return handler
}

// WithManagedReader bật API sổ cái ví cho Admin (chỉ chuyên gia mình quản lý).
func (h *Handler) WithManagedReader(reader *wallet.ManagedReader) *Handler {
	h.managed = reader
	return h
}
