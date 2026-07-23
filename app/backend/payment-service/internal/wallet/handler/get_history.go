package handler

import (
	"net/http"
	"payment-service/internal/payment/application/readquery"
	"payment-service/internal/wallet"
	"payment-service/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TxResponse struct {
	ID             uuid.UUID `json:"id"`
	WalletID       uuid.UUID `json:"wallet_id"`
	Type           string    `json:"type"`
	Amount         int64     `json:"amount"`
	BalanceAfter   int64     `json:"balance_after"`
	ReferenceType  string    `json:"reference_type"`
	ReferenceID    uuid.UUID `json:"reference_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      int64     `json:"created_at"`
}

// GetHistory handles GET /api/v1/payments/wallets/history.
// @Summary [PATIENT/EXPERT] Get paginated wallet history
// @Tags Wallets
// @Security BearerAuth
// @Param type query string false "Transaction type"
// @Param direction query string false "CREDIT or DEBIT"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /payments/wallets/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	role := c.GetHeader("X-User-Role")
	if role != "PATIENT" && role != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only patients or experts can view transaction history", "forbidden")
		return
	}
	userID, err := uuid.Parse(c.GetHeader("X-User-Id"))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}
	page, err := readquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid wallet history filter", err.Error())
		return
	}
	dateRange, err := readquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid wallet history filter", err.Error())
		return
	}
	txType, err := wallet.ParseTransactionType(c.Query("type"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid wallet history filter", err.Error())
		return
	}
	direction, err := wallet.ParseDirection(c.Query("direction"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid wallet history filter", err.Error())
		return
	}
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve transaction history", "wallet reader unavailable")
		return
	}
	result, err := h.reader.ListTransactionHistory(c.Request.Context(), wallet.TransactionHistoryQuery{UserID: userID, Type: txType, Direction: direction, FromMs: dateRange.FromMs, ToMs: dateRange.ToMs, Page: page})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve transaction history", err.Error())
		return
	}
	items := make([]TxResponse, len(result.Items))
	for i, tx := range result.Items {
		items[i] = TxResponse{ID: tx.ID, WalletID: tx.WalletID, Type: tx.Type.String(), Amount: tx.Amount.Int64(), BalanceAfter: tx.BalanceAfter.Int64(), ReferenceType: tx.ReferenceType, ReferenceID: tx.ReferenceID, IdempotencyKey: tx.IdempotencyKey, CreatedAt: tx.CreatedAt}
	}
	response.Success(c, "Transaction history retrieved successfully", gin.H{"items": items, "transactions": items, "page": result.Page, "size": result.Size, "total_items": result.TotalItems, "total_pages": result.TotalPages, "has_next": result.HasNext, "has_previous": result.HasPrevious})
}
