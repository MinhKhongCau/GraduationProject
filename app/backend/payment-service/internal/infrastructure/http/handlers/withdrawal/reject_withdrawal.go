package withdrawal

import (
	"net/http"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RejectWithdrawal handles POST /api/v1/payments/withdrawals/:id/reject
// @Summary      [ADMIN] Reject withdrawal request
// @Description  Admin rejects a pending withdrawal request. Locked funds are returned to the expert's wallet. Requires ADMIN role.
// @Tags         Admin - Withdrawals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string               true  "Mã yêu cầu rút tiền (UUID)"
// @Param        body         body      AdminProcessRequest  true  "Lý do từ chối (bắt buộc)"
// @Success      200          {object}  response.Response
// @Failure      400          {object}  response.Response
// @Failure      403          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/withdrawals/{id}/reject [post]
func (h *Handler) RejectWithdrawal(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Only administrators can reject withdrawals", "forbidden")
		return
	}

	adminIDStr := c.GetHeader("X-User-Id")
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}

	reqIDStr := c.Param("id")
	reqUUID, err := uuid.Parse(reqIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request ID format", err.Error())
		return
	}

	var req AdminProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Note == "" {
		response.Error(c, http.StatusBadRequest, "Note is required for rejections", "missing note")
		return
	}

	err = h.usecase.RejectWithdrawal(c.Request.Context(), adminID, reqUUID, req.Note)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reject withdrawal request", err.Error())
		return
	}

	response.Success(c, "Withdrawal request rejected successfully", nil)
}
