package wallet

import (
	"fmt"
	"net/http"
	"payment-service/internal/domain/money"
	"payment-service/internal/infrastructure/http/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
// @Security     BearerAuth
// @Param        body       body      TopUpRequest  true  "Thông tin nạp tiền"
// @Success      200        {object}  response.Response
// @Failure      400        {object}  response.Response
// @Failure      401        {object}  response.Response
// @Failure      500        {object}  response.Response
// @Router       /payments/wallets/top-up [post]
func (h *Handler) TopUpWallet(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" && userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only patients or experts can top up wallets", "forbidden")
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

	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid body", err.Error())
		return
	}

	idempotencyKey := fmt.Sprintf("topup_%s_%d", userID.String(), time.Now().UnixNano())

	err = h.usecase.CreditAvailable(c.Request.Context(), userID, money.Money(req.Amount), "MANUAL", uuid.New(), idempotencyKey)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to top up wallet", err.Error())
		return
	}

	response.Success(c, "Wallet topped up successfully", nil)
}
