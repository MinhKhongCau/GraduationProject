package handler

import (
	"fmt"
	"net/http"
	"payment-service/internal/domain/vo"
	"payment-service/internal/wallet"
	"payment-service/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	usecase wallet.Usecase
}

func NewHandler(usecase wallet.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

// GetWallet handles GET /api/v1/payments/wallets/me
// @Summary      [PATIENT/EXPERT] Get current user's wallet
// @Description  Retrieve wallet information (available balance, pending balance, locked balance) for the current user. Requires PATIENT or EXPERT role.
// @Tags         Wallets
// @Accept       json
// @Produce      json
// @Param        X-User-Id    header    string  true  "User ID (UUID)"
// @Param        X-User-Role  header    string  true  "User Role (PATIENT hoặc EXPERT)"
// @Success      200          {object}  response.Response
// @Failure      401          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/wallets/me [get]
func (h *Handler) GetWallet(c *gin.Context) {
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

	wallet, err := h.usecase.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve wallet", err.Error())
		return
	}

	response.Success(c, "Wallet retrieved successfully", wallet)
}

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
// @Param        X-User-Id    header    string  true  "User ID (UUID)"
// @Param        X-User-Role  header    string  true  "User Role (PATIENT hoặc EXPERT)"
// @Success      200          {object}  response.Response{data=[]TxResponse}
// @Failure      401          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/wallets/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
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

// TopUpRequest for development/testing
type TopUpRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

// TopUpWallet handles POST /api/v1/payments/wallets/top-up
// @Summary      [PATIENT/EXPERT] Top up wallet (Dev/Test only)
// @Description  Directly top up the available balance of the current user's wallet. Requires PATIENT or EXPERT role.
// @Tags         Wallets
// @Accept       json
// @Produce      json
// @Param        X-User-Id  header    string        true  "User ID (UUID)"
// @Param        body       body      TopUpRequest  true  "Thông tin nạp tiền"
// @Success      200        {object}  response.Response
// @Failure      400        {object}  response.Response
// @Failure      401        {object}  response.Response
// @Failure      500        {object}  response.Response
// @Router       /payments/wallets/top-up [post]
func (h *Handler) TopUpWallet(c *gin.Context) {
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

	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid body", err.Error())
		return
	}

	idempotencyKey := fmt.Sprintf("topup_%s_%d", userID.String(), time.Now().UnixNano())

	err = h.usecase.CreditAvailable(c.Request.Context(), userID, vo.Money(req.Amount), "MANUAL", uuid.New(), idempotencyKey)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to top up wallet", err.Error())
		return
	}

	response.Success(c, "Wallet topped up successfully", nil)
}
