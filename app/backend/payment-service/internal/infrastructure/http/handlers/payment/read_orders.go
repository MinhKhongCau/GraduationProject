package payment

import (
	"errors"
	"net/http"
	"payment-service/internal/application/managedscope"
	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"
	"payment-service/internal/infrastructure/http/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListPaymentOrders handles GET /api/v1/payments/orders.
// @Summary [PATIENT] List own payment orders
// @Tags Payments
// @Security BearerAuth
// @Param appointment_id query string false "Appointment UUID" default("dddddddd-dddd-4ddd-8ddd-dddddddddddd") example("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
// @Param status query string false "PENDING, SUCCESS, FAILED, EXPIRED" default("PENDING") example("PENDING")
// @Param type query string false "APPOINTMENT or TOP_UP"
// @Param fulfillment_status query string false "Fulfillment status"
// @Param from query string false "Start date YYYY-MM-DD" default("2026-07-01") example("2026-07-01")
// @Param to query string false "End date YYYY-MM-DD" default("2026-07-30") example("2026-07-30")
// @Param page query int false "Zero-based page" default(0) example(0)
// @Param size query int false "Page size, 1-100" default(20) example(20)
// @Router /payments/orders [get]
func (h *Handler) ListPaymentOrders(c *gin.Context) {
	payerID, ok := requireActor(c, "PATIENT", "Only patients can list payment orders")
	if !ok {
		return
	}
	filter, err := paymentOrderFilter(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment order filter", err.Error())
		return
	}
	filter.PayerID = payerID
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

// SummarizePaymentOrders handles GET /api/v1/payments/orders/summary.
// @Summary [PATIENT] Summarize own payment orders (total paid)
// @Tags Payments
// @Security BearerAuth
// @Param status query string false "PENDING, SUCCESS, FAILED, EXPIRED"
// @Param type query string false "APPOINTMENT or TOP_UP"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Router /payments/orders/summary [get]
func (h *Handler) SummarizePaymentOrders(c *gin.Context) {
	payerID, ok := requireActor(c, "PATIENT", "Only patients can summarize payment orders")
	if !ok {
		return
	}
	filter, err := paymentOrderFilter(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment order filter", err.Error())
		return
	}
	filter.PayerID = payerID
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to summarize payment orders", "payment order reader unavailable")
		return
	}
	result, err := h.reader.SummarizePatientOrders(c.Request.Context(), filter)
	if err != nil {
		writePaymentReadError(c, err, "Failed to summarize payment orders")
		return
	}
	response.Success(c, "Payment order summary retrieved successfully", result)
}

// GetPaymentOrder handles GET /api/v1/payments/orders/:id.
// @Summary [PATIENT] Get own payment order status
// @Tags Payments
// @Security BearerAuth
// @Param id path string true "Payment order UUID"
// @Router /payments/orders/{id} [get]
func (h *Handler) GetPaymentOrder(c *gin.Context) {
	payerID, ok := requireActor(c, "PATIENT", "Payment order access denied")
	if !ok {
		return
	}
	orderID, ok := pathOrderID(c)
	if !ok {
		return
	}
	if h.reader == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get payment order", "payment order reader unavailable")
		return
	}
	result, err := h.reader.GetPaymentOrder(c.Request.Context(), payerID, orderID)
	if err != nil {
		writePaymentReadError(c, err, "Failed to get payment order")
		return
	}
	response.Success(c, "Payment order retrieved successfully", result)
}

// requireActor kiểm tra vai trò (header X-User-Role do Gateway chèn) và trả về X-User-Id.
func requireActor(c *gin.Context, role, forbiddenMessage string) (uuid.UUID, bool) {
	if c.GetHeader("X-User-Role") != role {
		response.Error(c, http.StatusForbidden, forbiddenMessage, "forbidden")
		return uuid.Nil, false
	}
	actorID, err := uuid.Parse(strings.TrimSpace(c.GetHeader("X-User-Id")))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return uuid.Nil, false
	}
	return actorID, true
}

func pathOrderID(c *gin.Context) (uuid.UUID, bool) {
	orderID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment order ID", err.Error())
		return uuid.Nil, false
	}
	return orderID, true
}

// paymentOrderFilter đọc các tiêu chí lọc chung từ query; phạm vi (payer/expert) do handler điền.
func paymentOrderFilter(c *gin.Context) (apppayment.PaymentOrderFilter, error) {
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
	orderType, err := apppayment.ParsePaymentOrderType(c.Query("type"))
	if err != nil {
		return apppayment.PaymentOrderFilter{}, err
	}
	fulfillment, err := apppayment.ParseFulfillmentStatus(c.Query("fulfillment_status"))
	if err != nil {
		return apppayment.PaymentOrderFilter{}, err
	}
	filter := apppayment.PaymentOrderFilter{Status: status, Type: orderType, FulfillmentStatus: fulfillment, FromMs: dateRange.FromMs, ToMs: dateRange.ToMs, Page: page}
	if filter.AppointmentID, err = optionalUUID(c.Query("appointment_id")); err != nil {
		return filter, err
	}
	return filter, nil
}

func writePaymentReadError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, apppayment.ErrPaymentOrderNotFound):
		response.Error(c, http.StatusNotFound, "Payment order not found", err.Error())
	case errors.Is(err, apppayment.ErrPaymentOrderForbidden):
		response.Error(c, http.StatusForbidden, "Payment order access denied", err.Error())
	case errors.Is(err, managedscope.ErrNotManagedExpert):
		response.Error(c, http.StatusForbidden, "Expert is not managed by this administrator", err.Error())
	case errors.Is(err, managedscope.ErrResolverUnavailable):
		response.Error(c, http.StatusServiceUnavailable, "Cannot determine managed experts right now", err.Error())
	case errors.Is(err, apppayment.ErrCompensationCaseNotFound):
		response.Error(c, http.StatusNotFound, "Compensation case not found", err.Error())
	case errors.Is(err, apppayment.ErrCompensationCaseExists):
		response.Error(c, http.StatusConflict, "Compensation case already exists", err.Error())
	case errors.Is(err, paymentdomain.ErrInvalidAdminReviewAction), errors.Is(err, paymentdomain.ErrResolutionNoteTooLong):
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
	case errors.Is(err, paymentdomain.ErrOrderNotReviewable), errors.Is(err, paymentdomain.ErrCompensationAlreadyClosed):
		response.Error(c, http.StatusConflict, "Action not allowed in current state", err.Error())
	case errors.Is(err, apppayment.ErrInvalidCompensationFilter):
		response.Error(c, http.StatusBadRequest, "Invalid compensation case filter", err.Error())
	case errors.Is(err, apppayment.ErrInvalidPaymentOrderFilter), errors.Is(err, readquery.ErrInvalid):
		response.Error(c, http.StatusBadRequest, "Invalid payment order filter", err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, fallback, err.Error())
	}
}
