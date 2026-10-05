package withdrawal

import (
	"net/http"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetBankAccounts handles GET /api/v1/payments/bank-accounts
// @Summary      [EXPERT] Get linked bank accounts
// @Description  Retrieve all linked bank accounts for the current expert. Requires EXPERT role.
// @Tags         Withdrawals & Bank Accounts
// @Security     BearerAuth
// @Success      200          {object}  response.Response
// @Failure      401          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/bank-accounts [get]
func (h *Handler) GetBankAccounts(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only experts can get linked bank accounts", "forbidden")
		return
	}

	userIDStr := c.GetHeader("X-User-Id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}

	accounts, err := h.usecase.GetBankAccounts(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch bank accounts", err.Error())
		return
	}

	response.Success(c, "Bank accounts retrieved successfully", accounts)
}
