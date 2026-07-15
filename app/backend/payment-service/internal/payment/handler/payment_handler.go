package handler

import (
	"net/http"
	"payment-service/internal/domain/vo"
	"payment-service/internal/payment"
	"payment-service/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	usecase payment.Usecase
}

func NewHandler(usecase payment.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

type CreateOrderRequest struct {
	PayerID  string `json:"payer_id" binding:"required"`
	ExpertID string `json:"expert_id" binding:"required"`
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Gateway  string `json:"gateway" binding:"required"` // VNPAY | MOMO | MOCK
}

type CreateOrderResponse struct {
	OrderID          uuid.UUID `json:"order_id"`
	GrossAmount      int64     `json:"gross_amount"`
	NetAmount        int64     `json:"net_amount"`
	CommissionAmount int64     `json:"commission_amount"`
	PaymentURL       string    `json:"payment_url"`
	Status           string    `json:"status"`
}

// CreateOrder handles POST /api/v1/payments/orders
// @Summary      Tạo đơn hàng thanh toán mới
// @Description  Khởi tạo đơn hàng thanh toán cho cuộc hẹn. Trả về đường link thanh toán (Ví dụ: VNPay checkout). Yêu cầu role: PATIENT hoặc gọi nội bộ từ Booking Service.
// @Tags         Thanh toán (Payment Orders)
// @Accept       json
// @Produce      json
// @Param        body  body      CreateOrderRequest  true  "Thông tin khởi tạo giao dịch thanh toán"
// @Success      200   {object}  response.Response{data=CreateOrderResponse}
// @Failure      400   {object}  response.Response
// @Failure      500   {object}  response.Response
// @Router       /payments/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	payerUUID, err := uuid.Parse(req.PayerID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Payer ID format", err.Error())
		return
	}

	expertUUID, err := uuid.Parse(req.ExpertID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Expert ID format", err.Error())
		return
	}

	gatewayName := strings.ToUpper(req.Gateway)
	if gatewayName != "VNPAY" && gatewayName != "MOMO" && gatewayName != "MOCK" {
		response.Error(c, http.StatusBadRequest, "Unsupported gateway type", "Gateway must be VNPAY, MOMO, or MOCK")
		return
	}

	ipAddr := c.ClientIP()
	order, payURL, err := h.usecase.CreateOrder(c.Request.Context(), payerUUID, expertUUID, vo.Money(req.Amount), gatewayName, ipAddr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create payment order", err.Error())
		return
	}

	res := CreateOrderResponse{
		OrderID:          order.ID,
		GrossAmount:      order.GrossAmount.Int64(),
		NetAmount:        order.NetAmount.Int64(),
		CommissionAmount: order.CommissionAmount.Int64(),
		PaymentURL:       payURL,
		Status:           order.Status.String(),
	}

	response.Success(c, "Payment order created successfully", res)
}

// HandleVNPayIPN handles GET /api/v1/payments/vnpay-ipn
// @Summary      Webhook nhận thông báo kết quả thanh toán từ VNPay (IPN)
// @Description  Cổng thanh toán VNPay gọi API này để cập nhật trạng thái thanh toán của đơn hàng. Yêu cầu role: Public (không cần token, kiểm tra bằng mã checksum).
// @Tags         Thanh toán (Payment Orders)
// @Produce      json
// @Param        queryParams  query     interface{}  false  "Các tham số phản hồi tự động từ VNPay"
// @Success      200          {object}  map[string]string
// @Router       /payments/vnpay-ipn [get]
func (h *Handler) HandleVNPayIPN(c *gin.Context) {
	queryParams := c.Request.URL.Query()

	alreadyProcessed, err := h.usecase.ProcessIPN(c.Request.Context(), queryParams)
	if err != nil {
		logStr := err.Error()
		if strings.Contains(logStr, "checksum") {
			c.JSON(http.StatusOK, gin.H{"RspCode": "97", "Message": "Invalid Signature"})
			return
		}
		if strings.Contains(logStr, "not found") {
			c.JSON(http.StatusOK, gin.H{"RspCode": "01", "Message": "Order not found"})
			return
		}
		if strings.Contains(logStr, "amount") {
			c.JSON(http.StatusOK, gin.H{"RspCode": "04", "Message": "Invalid amount"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"RspCode": "99", "Message": "Unknown error: " + logStr})
		return
	}

	if alreadyProcessed {
		c.JSON(http.StatusOK, gin.H{"RspCode": "02", "Message": "Order already confirmed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"RspCode": "00", "Message": "Confirm Success"})
}
