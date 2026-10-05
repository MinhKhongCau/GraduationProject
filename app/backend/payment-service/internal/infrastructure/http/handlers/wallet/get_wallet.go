package wallet

import (
	"net/http"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetWallet handles GET /api/v1/payments/wallets/me
// @Summary      [PATIENT/EXPERT] Get current user's wallet
// @Description  Retrieve wallet information (available balance, pending balance, locked balance) for the current user. Requires PATIENT or EXPERT role.
// @Tags         Wallets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200          {object}  response.Response
// @Failure      401          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/wallets/me [get]
func (h *Handler) GetWallet(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "PATIENT" && userRole != "EXPERT" {
		response.Error(c, http.StatusForbidden, "Only patients or experts can view wallets", "forbidden")
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

	wallet, err := h.usecase.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve wallet", err.Error())
		return
	}

	response.Success(c, "Wallet retrieved successfully", wallet)
}
