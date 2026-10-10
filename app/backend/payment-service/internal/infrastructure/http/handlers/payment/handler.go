package payment

import (
	apppayment "payment-service/internal/application/payment"
)

type Handler struct {
	usecase apppayment.Usecase
	reader  apppayment.ReadUsecase
	manager apppayment.ManageUsecase
	// settlement chi trả thù lao chuyên gia khi buổi tư vấn hoàn tất (route nội bộ).
	settlement apppayment.SettlementUsecase
}

func NewHandler(usecase apppayment.Usecase) *Handler {
	handler := &Handler{usecase: usecase}
	if reader, ok := usecase.(apppayment.ReadUsecase); ok {
		handler.reader = reader
	}
	if manager, ok := usecase.(apppayment.ManageUsecase); ok {
		handler.manager = manager
	}
	if settlement, ok := usecase.(apppayment.SettlementUsecase); ok {
		handler.settlement = settlement
	}
	return handler
}
