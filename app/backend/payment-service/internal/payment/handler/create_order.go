package handler

import (
	"net/http"
	"payment-service/internal/domain/vo"
	"payment-service/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateOrderRequest struct {
	PayerID        string  `json:"payer_id" binding:"required"`
	ExpertID       string  `json:"expert_id" binding:"required"`
	Amount         int64   `json:"amount" binding:"required,gt=0"`
	Gateway        string  `json:"gateway" binding:"required"` // VNPAY | MOMO | MOCK
	// AppointmentID liên kết order này với lịch hẹn. Optional — nếu không cung cấp
	// thì đây là order nạp tiền ví trực tiếp, không liên quan đến booking.
	AppointmentID  *string `json:"appointment_id,omitempty"`
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
// @Summary      [PATIENT/SYSTEM] Create a new payment order
// @Description  Create a payment order for an appointment. Returns the payment URL (e.g. VNPay checkout). Requires PATIENT role or internal call from Booking Service.
// @Tags         Payment Orders
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

	// Parse appointment_id (optional)
	var appointmentID *string
	if req.AppointmentID != nil && *req.AppointmentID != "" {
		appointmentID = req.AppointmentID
	}

	ipAddr := c.ClientIP()
	order, payURL, err := h.usecase.CreateOrder(c.Request.Context(), payerUUID, expertUUID, vo.Money(req.Amount), gatewayName, ipAddr, appointmentID)
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
