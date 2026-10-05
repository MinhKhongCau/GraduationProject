package withdrawal

import (
	"net/http"
	"payment-service/internal/domain/money"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateWithdrawalRequest struct {
	BankAccountID string `json:"bank_account_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
}

type WithdrawalResponse struct {
	ID                     uuid.UUID `json:"id"`
	Amount                 int64     `json:"amount"`
	Status                 string    `json:"status"`
	RequiresManualApproval bool      `json:"requires_manual_approval"`
	RequestedAt            int64     `json:"requested_at"`
}

// CreateWithdrawal handles POST /api/v1/payments/withdrawals
// @Summary      [EXPERT] Request a withdrawal
// @Description  Request a withdrawal to a linked bank account. Amounts < 5,000,000 VND are auto-approved. Requires EXPERT role.
// @Tags         Withdrawals & Bank Accounts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body         body      CreateWithdrawalRequest  true  "Thông tin yêu cầu rút tiền"
// @Success      200          {object}  response.Response{data=WithdrawalResponse}
// @Failure      400          {object}  response.Response
// @Failure      401          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/withdrawals [post]
func (h *Handler) CreateWithdrawal(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can request withdrawals", "forbidden")
		return
	}

	userIDStr := c.GetHeader("X-User-Id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}

	var req CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	bankAccountUUID, err := uuid.Parse(req.BankAccountID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Bank Account ID format", err.Error())
		return
	}

	request, err := h.usecase.CreateWithdrawal(c.Request.Context(), userID, bankAccountUUID, money.Money(req.Amount))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to request withdrawal", err.Error())
		return
	}

	res := WithdrawalResponse{
		ID:                     request.ID,
		Amount:                 request.Amount.Int64(),
		Status:                 request.Status.String(),
		RequiresManualApproval: request.RequiresManualApproval,
		RequestedAt:            request.RequestedAt,
	}

	response.Success(c, "Withdrawal request created successfully", res)
}
