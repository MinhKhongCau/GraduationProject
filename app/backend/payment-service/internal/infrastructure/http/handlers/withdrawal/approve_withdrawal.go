package withdrawal

import (
	"net/http"
	"payment-service/internal/infrastructure/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminProcessRequest struct {
	Note string `json:"note"`
}

// ApproveWithdrawal handles POST /api/v1/payments/withdrawals/:id/approve
// @Summary      [ADMIN] Approve withdrawal request
// @Description  Admin approves a pending withdrawal request. Funds are transferred via Payout Gateway. Requires ADMIN role.
// @Tags         Admin - Withdrawals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string               true  "Mã yêu cầu rút tiền (UUID)"
// @Param        body         body      AdminProcessRequest  true  "Thông tin ghi chú duyệt"
// @Success      200          {object}  response.Response
// @Failure      400          {object}  response.Response
// @Failure      403          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/withdrawals/{id}/approve [post]
func (h *Handler) ApproveWithdrawal(c *gin.Context) {
	userRole := c.GetHeader("X-User-Role")
	if userRole != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Only administrators can approve withdrawals", "forbidden")
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
	_ = c.ShouldBindJSON(&req) // Ghi chú là không bắt buộc khi approve

	err = h.usecase.ApproveWithdrawal(c.Request.Context(), adminID, reqUUID, req.Note)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to approve withdrawal request", err.Error())
		return
	}

	response.Success(c, "Withdrawal request approved successfully", nil)
}
