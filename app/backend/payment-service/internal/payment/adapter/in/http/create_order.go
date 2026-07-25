package handler

import (
	"errors"
	"net/http"
	apppayment "payment-service/internal/payment/application"
	"payment-service/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateOrderRequest struct {
	// Booking appointment ID to pay with VNPay.
	AppointmentID string `json:"appointment_id" binding:"required" example:"dddddddd-dddd-4ddd-8ddd-dddddddddddd"`
}

type CreateOrderResponse struct {
	OrderID          uuid.UUID `json:"order_id"`
	GrossAmount      int64     `json:"gross_amount"`
	NetAmount        int64     `json:"net_amount"`
	CommissionAmount int64     `json:"commission_amount"`
	PaymentURL       string    `json:"payment_url"`
	Status           string    `json:"status"`
	ExpiresAt        int64     `json:"expires_at"`
}

// CreateOrder handles POST /api/v1/payments/orders.
// @Summary      [PATIENT/SYSTEM] Create a new payment order
// @Description  Create a VNPay payment order for a booking appointment. Payer comes from X-User-Id; expert and amount come from Booking Service.
// @Tags         Payment Orders
// @Accept       json
// @Produce      json
// @Param        body  body      CreateOrderRequest  true  "Appointment payment order payload"
// @Success      200   {object}  response.Response{data=CreateOrderResponse}
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Failure      403   {object}  response.Response
// @Failure      404   {object}  response.Response
// @Failure      409   {object}  response.Response
// @Failure      500   {object}  response.Response
// @Security     BearerAuth
// @Router       /payments/orders [post]
// @Param        X-User-Id  header  string  true  "User ID populated by API Gateway from token"
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	payerID := strings.TrimSpace(c.GetHeader("X-User-Id"))
	if payerID == "" {
		response.Error(c, http.StatusUnauthorized, "Missing authenticated payer", "X-User-Id header is required")
		return
	}

	payerUUID, err := uuid.Parse(payerID)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid authenticated payer", err.Error())
		return
	}

	appointmentID := strings.TrimSpace(req.AppointmentID)
	if appointmentID == "" {
		response.Error(c, http.StatusBadRequest, "Invalid payment order request", "appointment_id is required")
		return
	}

	order, payURL, err := h.usecase.CreateOrder(
		c.Request.Context(),
		payerUUID,
		appointmentID,
		c.ClientIP(),
	)
	if err != nil {
		status, message := createOrderErrorResponse(err)
		response.Error(c, status, message, err.Error())
		return
	}

	res := CreateOrderResponse{
		OrderID:          order.ID,
		GrossAmount:      order.GrossAmount.Int64(),
		NetAmount:        order.NetAmount.Int64(),
		CommissionAmount: order.CommissionAmount.Int64(),
		PaymentURL:       payURL,
		Status:           order.Status.String(),
		ExpiresAt:        order.ExpiresAt,
	}

	response.Success(c, "Payment order created successfully", res)
}

func createOrderErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, apppayment.ErrUnsupportedGateway),
		errors.Is(err, apppayment.ErrInvalidCreateOrderRequest),
		errors.Is(err, apppayment.ErrInvalidBookingData):
		return http.StatusBadRequest, "Invalid payment order request"
	case errors.Is(err, apppayment.ErrAppointmentOwnership):
		return http.StatusForbidden, "Appointment does not belong to payer"
	case errors.Is(err, apppayment.ErrBookingAppointmentNotFound):
		return http.StatusNotFound, "Appointment not found"
	case errors.Is(err, apppayment.ErrAppointmentInvalidState):
		return http.StatusConflict, "Appointment is not payable"
	case errors.Is(err, apppayment.ErrAppointmentAlreadyPaid):
		return http.StatusConflict, "Appointment is already paid"
	case errors.Is(err, apppayment.ErrPaymentWindowTooShort):
		return http.StatusConflict, "Payment window has expired"
	case errors.Is(err, apppayment.ErrExistingOrderConflict):
		return http.StatusConflict, "Existing payment order cannot be reused"
	case errors.Is(err, apppayment.ErrActivePendingOrderExists):
		return http.StatusConflict, "Active payment order already exists"
	default:
		return http.StatusInternalServerError, "Failed to create payment order"
	}
}
