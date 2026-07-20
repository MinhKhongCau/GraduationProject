package handler

import (
	"errors"
	"net/http"
	apppayment "payment-service/internal/payment/application"
	"payment-service/internal/payment/application/readquery"
	"payment-service/pkg/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListPaymentOrders handles GET /api/v1/payments/orders.
// @Summary [PATIENT] List own payment orders
// @Tags Payments
// @Security BearerAuth
// @Param appointment_id query string false "Appointment UUID"
// @Param status query string false "PENDING, SUCCESS, FAILED, EXPIRED"
// @Param fulfillment_status query string false "Fulfillment status"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /payments/orders [get]
func (h *Handler) ListPaymentOrders(c *gin.Context) {
	if c.GetHeader("X-User-Role") != "PATIENT" {
		response.Error(c, http.StatusForbidden, "Only patients can list payment orders", "forbidden")
		return
	}
	payerID, err := uuid.Parse(strings.TrimSpace(c.GetHeader("X-User-Id")))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}
	filter, err := paymentOrderFilter(c, payerID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment order filter", err.Error())
		return
	}
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list payment orders", "payment order reader unavailable")
		return
	}
	result, err := h.reader.ListPaymentOrders(c.Request.Context(), filter)
	if err != nil {
		writePaymentReadError(c, err, "Failed to list payment orders")
		return
	}
	response.Success(c, "Payment orders retrieved successfully", result)
}

// GetPaymentOrder handles GET /api/v1/payments/orders/:id.
// @Summary [PATIENT/ADMIN] Get payment order status
// @Tags Payments
// @Security BearerAuth
// @Param id path string true "Payment order UUID"
// @Router /payments/orders/{id} [get]
func (h *Handler) GetPaymentOrder(c *gin.Context) {
	role := c.GetHeader("X-User-Role")
	if role != "PATIENT" && role != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Payment order access denied", "forbidden")
		return
	}
	actorID, err := uuid.Parse(strings.TrimSpace(c.GetHeader("X-User-Id")))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return
	}
	orderID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment order ID", err.Error())
		return
	}
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get payment order", "payment order reader unavailable")
		return
	}
	result, err := h.reader.GetPaymentOrder(c.Request.Context(), actorID, orderID, role == "ADMIN")
	if err != nil {
		writePaymentReadError(c, err, "Failed to get payment order")
		return
	}
	response.Success(c, "Payment order retrieved successfully", result)
}

func paymentOrderFilter(c *gin.Context, payerID uuid.UUID) (apppayment.PaymentOrderFilter, error) {
	page, err := readquery.ParsePage(c.Query("page"), c.Query("size"))
	if err != nil {
		return apppayment.PaymentOrderFilter{}, err
	}
	dateRange, err := readquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		return apppayment.PaymentOrderFilter{}, err
	}
	status, err := apppayment.ParsePaymentOrderStatus(c.Query("status"))
	if err != nil {
		return apppayment.PaymentOrderFilter{}, err
	}
	fulfillment, err := apppayment.ParseFulfillmentStatus(c.Query("fulfillment_status"))
	if err != nil {
		return apppayment.PaymentOrderFilter{}, err
	}
	filter := apppayment.PaymentOrderFilter{PayerID: payerID, Status: status, FulfillmentStatus: fulfillment, FromMs: dateRange.FromMs, ToMs: dateRange.ToMs, Page: page}
	if value := strings.TrimSpace(c.Query("appointment_id")); value != "" {
		parsed, parseErr := uuid.Parse(value)
		if parseErr != nil {
			return filter, parseErr
		}
		filter.AppointmentID = &parsed
	}
	return filter, nil
}

func writePaymentReadError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, apppayment.ErrPaymentOrderNotFound):
		response.Error(c, http.StatusNotFound, "Payment order not found", err.Error())
	case errors.Is(err, apppayment.ErrPaymentOrderForbidden):
		response.Error(c, http.StatusForbidden, "Payment order access denied", err.Error())
	case errors.Is(err, apppayment.ErrInvalidPaymentOrderFilter), errors.Is(err, readquery.ErrInvalid):
		response.Error(c, http.StatusBadRequest, "Invalid payment order filter", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, fallback, err.Error())
	}
}
