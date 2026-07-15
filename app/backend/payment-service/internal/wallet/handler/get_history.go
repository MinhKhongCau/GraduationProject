package handler

import (
	"net/http"
	"payment-service/pkg/response"

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

// GetHistory handles GET /api/v1/payments/wallets/history
// @Summary      [PATIENT/EXPERT] Get transaction history
// @Description  Retrieve the list of transactions (balance changes) for the current user's wallet, sorted by latest time. Requires PATIENT or EXPERT role.
// @Tags         Wallets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200          {object}  response.Response{data=[]TxResponse}
// @Failure      401          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/wallets/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" && userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only patients or experts can view transaction history", "forbidden")
		return
	}

	userIDStr := c.GetHeader("X-User-Id")
	if userIDStr == "" {
		response.Error(c, http.StatusUnauthorized, "Missing X-User-Id header", "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid User ID format", err.Error())
		return
	}

	txs, err := h.usecase.GetTransactionHistory(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve transaction history", err.Error())
		return
	}

	res := make([]TxResponse, len(txs))
	for i, tx := range txs {
		res[i] = TxResponse{
			ID:             tx.ID,
			WalletID:       tx.WalletID,
			Type:           tx.Type.String(),
			Amount:         tx.Amount.Int64(),
			BalanceAfter:   tx.BalanceAfter.Int64(),
			ReferenceType:  tx.ReferenceType,
			ReferenceID:    tx.ReferenceID,
			IdempotencyKey: tx.IdempotencyKey,
			CreatedAt:      tx.CreatedAt,
		}
	}

	response.Success(c, "Transaction history retrieved successfully", res)
}
