package withdrawal

import (
	"errors"
	"net/http"
	"payment-service/internal/application/readquery"
	"payment-service/internal/application/withdrawal"
	"payment-service/internal/infrastructure/http/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListWithdrawals handles GET /api/v1/payments/withdrawals.
// @Summary [EXPERT] List own withdrawals
// @Tags Withdrawals
// @Security BearerAuth
// @Param status query string false "Withdrawal status"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /payments/withdrawals [get]
func (h *Handler) ListWithdrawals(c *gin.Context) { h.listWithdrawals(c, false) }

// ListAdminWithdrawals handles GET /api/v1/payments/admin/withdrawals.
// @Summary [ADMIN] List withdrawals
// @Tags Admin - Withdrawals
// @Security BearerAuth
// @Param expert_id query string false "Expert UUID"
// @Param status query string false "Withdrawal status"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /payments/admin/withdrawals [get]
func (h *Handler) ListAdminWithdrawals(c *gin.Context) { h.listWithdrawals(c, true) }

func (h *Handler) listWithdrawals(c *gin.Context, admin bool) {
	role := c.GetHeader("X-User-Role")
	if (!admin && role != "EXPERT") || (admin && role != "ADMIN") {
		response.Error(c, http.StatusForbidden, "Withdrawal access denied", "forbidden")
		return
	}
	actorID, err := uuid.Parse(strings.TrimSpace(c.GetHeader("X-User-Id")))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}
	page, err := readquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid withdrawal filter", err.Error())
		return
	}
	dateRange, err := readquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid withdrawal filter", err.Error())
		return
	}
	status, err := withdrawal.ParseWithdrawalStatus(c.Query("status"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid withdrawal filter", err.Error())
		return
	}
	filter := withdrawal.WithdrawalFilter{ActorID: actorID, IsAdmin: admin, Status: status, FromMs: dateRange.FromMs, ToMs: dateRange.ToMs, Page: page}
	if admin && strings.TrimSpace(c.Query("expert_id")) != "" {
		parsed, parseErr := uuid.Parse(c.Query("expert_id"))
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "Invalid withdrawal filter", parseErr.Error())
			return
		}
		filter.ExpertID = &parsed
	}
	result, err := h.usecase.ListWithdrawals(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list withdrawals", err.Error())
		return
	}
	response.Success(c, "Withdrawals retrieved successfully", result)
}

// GetWithdrawal handles GET /api/v1/payments/withdrawals/:id.
// @Summary [EXPERT/ADMIN] Get withdrawal detail
// @Tags Withdrawals
// @Security BearerAuth
// @Param id path string true "Withdrawal UUID"
// @Router /payments/withdrawals/{id} [get]
func (h *Handler) GetWithdrawal(c *gin.Context) {
	role := c.GetHeader("X-User-Role")
	if role != "EXPERT" && role != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Withdrawal access denied", "forbidden")
		return
	}
	actorID, err := uuid.Parse(strings.TrimSpace(c.GetHeader("X-User-Id")))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid withdrawal ID", err.Error())
		return
	}
	result, err := h.usecase.GetWithdrawal(c.Request.Context(), actorID, requestID, role == "ADMIN")
	if err != nil {
		switch {
		case errors.Is(err, withdrawal.ErrWithdrawalNotFound):
			response.Error(c, http.StatusNotFound, "Withdrawal not found", err.Error())
		case errors.Is(err, withdrawal.ErrWithdrawalForbidden):
			response.Error(c, http.StatusForbidden, "Withdrawal access denied", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to get withdrawal", err.Error())
		}
		return
	}
	response.Success(c, "Withdrawal retrieved successfully", result)
}
