package wallet

import (
	"errors"
	"net/http"
	"payment-service/internal/application/managedscope"
	"payment-service/internal/application/readquery"
	"payment-service/internal/application/wallet"
	"payment-service/internal/infrastructure/http/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ManagedTxResponse struct {
	TxResponse
	ExpertID uuid.UUID `json:"expert_id"`
}

// ListManagedTransactions handles GET /api/v1/payments/admin/wallet-transactions.
// @Summary [ADMIN] List wallet ledger of experts I manage
// @Tags Admin - Payments
// @Security BearerAuth
// @Param expert_id query string false "Expert auth UUID (must be managed by me)"
// @Param type query string false "PAYMENT_RECEIVED, COMMISSION_DEDUCTED, REFUND, WITHDRAWAL_LOCKED, WITHDRAWAL_COMPLETED, WITHDRAWAL_REJECTED, ADJUSTMENT"
// @Param direction query string false "CREDIT or DEBIT"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /payments/admin/wallet-transactions [get]
func (h *Handler) ListManagedTransactions(c *gin.Context) {
	if c.GetHeader("X-User-Role") != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Only administrators can view managed wallet transactions", "forbidden")
		return
	}
	adminID, err := uuid.Parse(strings.TrimSpace(c.GetHeader("X-User-Id")))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}
	query, err := managedTransactionQuery(c, adminID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid wallet transaction filter", err.Error())
		return
	}
	if h.managed == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve wallet transactions", "managed wallet reader unavailable")
		return
	}
	result, err := h.managed.ListManagedTransactions(c.Request.Context(), query)
	if err != nil {
		switch {
		case errors.Is(err, managedscope.ErrNotManagedExpert):
			response.Error(c, http.StatusForbidden, "Expert is not managed by this administrator", err.Error())
		case errors.Is(err, managedscope.ErrResolverUnavailable):
			response.Error(c, http.StatusServiceUnavailable, "Cannot determine managed experts right now", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve wallet transactions", err.Error())
		}
		return
	}
	items := make([]ManagedTxResponse, len(result.Items))
	for i, tx := range result.Items {
		items[i] = ManagedTxResponse{
			TxResponse: TxResponse{ID: tx.ID, WalletID: tx.WalletID, Type: tx.Type.String(), Amount: tx.Amount.Int64(), BalanceAfter: tx.BalanceAfter.Int64(), ReferenceType: tx.ReferenceType, ReferenceID: tx.ReferenceID, IdempotencyKey: tx.IdempotencyKey, CreatedAt: tx.CreatedAt},
			ExpertID:   tx.UserID,
		}
	}
	page := readquery.NewPage(items, readquery.PageRequest{Page: result.Page, Size: result.Size}, result.TotalItems)
	response.Success(c, "Wallet transactions retrieved successfully", page)
}

func managedTransactionQuery(c *gin.Context, adminID uuid.UUID) (wallet.ManagedTransactionQuery, error) {
	query := wallet.ManagedTransactionQuery{AdminID: adminID}
	var err error
	if query.Page, err = readquery.ParsePage(c.Query("page"), c.Query("size")); err != nil {
		return query, err
	}
	dateRange, err := readquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		return query, err
	}
	query.FromMs, query.ToMs = dateRange.FromMs, dateRange.ToMs
	if query.Type, err = wallet.ParseTransactionType(c.Query("type")); err != nil {
		return query, err
	}
	if query.Direction, err = wallet.ParseDirection(c.Query("direction")); err != nil {
		return query, err
	}
	if value := strings.TrimSpace(c.Query("expert_id")); value != "" {
		expertID, err := uuid.Parse(value)
		if err != nil {
			return query, err
		}
		query.ExpertID = &expertID
	}
	return query, nil
}
