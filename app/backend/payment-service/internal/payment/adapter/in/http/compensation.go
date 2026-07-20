package handler

import (
	"errors"
	"net/http"
	apppayment "payment-service/internal/payment/application"
	"payment-service/internal/payment/application/readquery"
	paymentdomain "payment-service/internal/payment/domain"
	"payment-service/pkg/response"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListCompensationCases handles GET /api/v1/payments/compensation-cases.
// @Summary      [ADMIN] List payment compensation cases
// @Tags         Admin - Payments
// @Produce      json
// @Security     BearerAuth
// @Param        status            query string false "MANUAL_REVIEW or REFUND_REQUIRED"
// @Param        appointment_id    query string false "Appointment UUID"
// @Param        payment_order_id  query string false "Payment order UUID"
// @Param        page              query int    false "Page number"
// @Param        size              query int    false "Page size, maximum 100"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /payments/compensation-cases [get]
func (h *Handler) ListCompensationCases(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	filter, err := compensationFilterFromRequest(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid compensation case filter", err.Error())
		return
	}
	result, err := h.usecase.ListCompensationCases(c.Request.Context(), filter)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to list compensation cases"
		if errors.Is(err, apppayment.ErrInvalidCompensationFilter) {
			status = http.StatusBadRequest
			message = "Invalid compensation case filter"
		}
		response.Error(c, status, message, err.Error())
		return
	}
	response.Success(c, "Compensation cases retrieved successfully", result)
}

// GetCompensationCase handles GET /api/v1/payments/compensation-cases/:id.
// @Summary      [ADMIN] Get a payment compensation case
// @Tags         Admin - Payments
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Compensation case UUID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /payments/compensation-cases/{id} [get]
func (h *Handler) GetCompensationCase(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	caseID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid compensation case ID", err.Error())
		return
	}
	result, err := h.usecase.GetCompensationCase(c.Request.Context(), caseID)
	if err != nil {
		switch {
		case errors.Is(err, apppayment.ErrCompensationCaseNotFound):
			response.Error(c, http.StatusNotFound, "Compensation case not found", err.Error())
		case errors.Is(err, apppayment.ErrInvalidCompensationFilter):
			response.Error(c, http.StatusBadRequest, "Invalid compensation case ID", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to get compensation case", err.Error())
		}
		return
	}
	response.Success(c, "Compensation case retrieved successfully", result)
}

func requireAdmin(c *gin.Context) bool {
	if c.GetHeader("X-User-Role") != "ADMIN" {
		response.Error(c, http.StatusForbidden, "Only administrators can view compensation cases", "forbidden")
		return false
	}
	if _, err := uuid.Parse(strings.TrimSpace(c.GetHeader("X-User-Id"))); err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or missing X-User-Id header", err.Error())
		return false
	}
	return true
}

func compensationFilterFromRequest(c *gin.Context) (apppayment.CompensationCaseFilter, error) {
	filter := apppayment.CompensationCaseFilter{Status: paymentdomain.CompensationStatus(strings.TrimSpace(c.Query("status"))), ReasonCode: paymentdomain.CompensationReasonCode(strings.TrimSpace(c.Query("reason_code")))}
	var err error
	if filter.AppointmentID, err = optionalUUID(c.Query("appointment_id")); err != nil {
		return filter, err
	}
	if filter.PaymentOrderID, err = optionalUUID(c.Query("payment_order_id")); err != nil {
		return filter, err
	}
	if value := strings.TrimSpace(c.Query("page")); value != "" {
		filter.Page, err = strconv.Atoi(value)
		if err != nil || filter.Page < 0 {
			return filter, errors.New("page must be an integer greater than or equal to 0")
		}
	}
	if value := strings.TrimSpace(c.Query("size")); value != "" {
		filter.Size, err = strconv.Atoi(value)
		if err != nil || filter.Size <= 0 {
			return filter, errors.New("size must be a positive integer")
		}
	}
	dateRange, err := readquery.ParseDateRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		return filter, err
	}
	filter.FromMs, filter.ToMs = dateRange.FromMs, dateRange.ToMs
	return filter, nil
}

func optionalUUID(value string) (*uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
