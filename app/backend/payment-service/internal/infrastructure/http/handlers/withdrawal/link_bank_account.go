package withdrawal

import (
	"net/http"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LinkBankAccountRequest struct {
	BankCode          string `json:"bank_code" binding:"required"`
	AccountNumber     string `json:"account_number" binding:"required"`
	AccountHolderName string `json:"account_holder_name" binding:"required"`
}

// LinkBankAccount handles POST /api/v1/payments/bank-accounts
// @Summary      [EXPERT] Link a new bank account
// @Description  Link a new bank account for withdrawals. Requires EXPERT role.
// @Tags         Withdrawals & Bank Accounts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body         body      LinkBankAccountRequest  true  "Thông tin tài khoản ngân hàng để liên kết"
// @Success      200          {object}  response.Response
// @Failure      400          {object}  response.Response
// @Failure      403          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/bank-accounts [post]
func (h *Handler) LinkBankAccount(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can link bank accounts", "forbidden")
		return
	}

	userIDStr := c.GetHeader("X-User-Id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}

	var req LinkBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	account, err := h.usecase.LinkBankAccount(c.Request.Context(), userID, req.BankCode, req.AccountNumber, req.AccountHolderName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to link bank account", err.Error())
		return
	}

	response.Success(c, "Bank account linked successfully", account)
}
