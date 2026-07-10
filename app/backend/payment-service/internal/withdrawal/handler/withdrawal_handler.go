package handler

import (
	"net/http"
	"payment-service/internal/domain"
	"payment-service/internal/withdrawal"
	"payment-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	usecase withdrawal.Usecase
}

func NewHandler(usecase withdrawal.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

type LinkBankAccountRequest struct {
	BankCode          string `json:"bank_code" binding:"required"`
	AccountNumber     string `json:"account_number" binding:"required"`
	AccountHolderName string `json:"account_holder_name" binding:"required"`
}

// LinkBankAccount handles POST /api/v1/payments/bank-accounts
// @Summary      Liên kết tài khoản ngân hàng mới
// @Description  Chuyên gia thực hiện liên kết tài khoản ngân hàng để chuẩn bị cho việc rút tiền từ ví. Yêu cầu role: EXPERT.
// @Tags         Rút tiền & Tài khoản ngân hàng
// @Accept       json
// @Produce      json
// @Param        X-User-Id    header    string                  true  "User ID (UUID)"
// @Param        X-User-Role  header    string                  true  "User Role (EXPERT)"
// @Param        body         body      LinkBankAccountRequest  true  "Thông tin tài khoản ngân hàng để liên kết"
// @Success      200          {object}  response.Response{data=domain.BankAccount}
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

// GetBankAccounts handles GET /api/v1/payments/bank-accounts
// @Summary      Lấy danh sách tài khoản ngân hàng đã liên kết
// @Description  Lấy toàn bộ các tài khoản ngân hàng đã liên kết của chuyên gia hiện tại. Yêu cầu role: EXPERT.
// @Tags         Rút tiền & Tài khoản ngân hàng
// @Produce      json
// @Param        X-User-Id    header    string  true  "User ID (UUID)"
// @Param        X-User-Role  header    string  true  "User Role (EXPERT)"
// @Success      200          {object}  response.Response{data=[]domain.BankAccount}
// @Failure      401          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /payments/bank-accounts [get]
func (h *Handler) GetBankAccounts(c *gin.Context) {
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
// @Summary      Yêu cầu rút tiền từ ví về tài khoản ngân hàng
// @Description  Chuyên gia tạo phiếu yêu cầu rút tiền khả dụng. Dưới 5,000,000 VND sẽ tự động duyệt và chuyển khoản, trên 5,000,000 VND sẽ cần ADMIN duyệt thủ công. Yêu cầu role: EXPERT.
// @Tags         Rút tiền & Tài khoản ngân hàng
// @Accept       json
// @Produce      json
// @Param        X-User-Id    header    string                   true  "User ID (UUID)"
// @Param        X-User-Role  header    string                   true  "User Role (EXPERT)"
// @Param        body         body      CreateWithdrawalRequest  true  "Thông tin yêu cầu rút tiền"
// @Success      200          {object}  response.Response{data=WithdrawalResponse}
// @Failure      400          {object}  response.Response
// @Failure      403          {object}  response.Response
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

	request, err := h.usecase.CreateWithdrawal(c.Request.Context(), userID, bankAccountUUID, domain.Money(req.Amount))
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

type AdminProcessRequest struct {
	Note string `json:"note"`
}

// ApproveWithdrawal handles POST /api/v1/payments/withdrawals/:id/approve
// @Summary      Duyệt yêu cầu rút tiền của chuyên gia (Dành cho Admin)
// @Description  Quản trị viên phê duyệt yêu cầu rút tiền. Tiền được chuyển khoản qua Payout Gateway. Yêu cầu role: ADMIN.
// @Tags         Rút tiền & Tài khoản ngân hàng (Quản trị viên)
// @Accept       json
// @Produce      json
// @Param        X-User-Id    header    string               true  "Admin ID (UUID)"
// @Param        X-User-Role  header    string               true  "Admin Role (ADMIN)"
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

// RejectWithdrawal handles POST /api/v1/payments/withdrawals/:id/reject
// @Summary      Từ chối yêu cầu rút tiền của chuyên gia (Dành cho Admin)
// @Description  Quản trị viên từ chối yêu cầu rút tiền. Tiền bị khóa (locked_balance) được mở khóa hoàn trả lại ví chuyên gia (available_balance). Yêu cầu role: ADMIN.
// @Tags         Rút tiền & Tài khoản ngân hàng (Quản trị viên)
// @Accept       json
// @Produce      json
// @Param        X-User-Id    header    string               true  "Admin ID (UUID)"
// @Param        X-User-Role  header    string               true  "Admin Role (ADMIN)"
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
