package payment

import (
	"errors"
	"net/http"
	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/application/readquery"
	paymentdomain "payment-service/internal/domain/payment"
	"payment-service/internal/infrastructure/http/response"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListCompensationCases handles GET /api/v1/payments/compensation-cases.
// @Summary      [ADMIN] List payment compensation cases of experts I manage
// @Tags         Admin - Payments
// @Produce      json
// @Security     BearerAuth
// @Param        status            query string false "MANUAL_REVIEW, REFUND_REQUIRED or RESOLVED"
// @Param        expert_id         query string false "Expert auth UUID (must be managed by me)"
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
	adminID, ok := requireAdmin(c)
	if !ok {
		return
	}
	filter, err := compensationFilterFromRequest(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid compensation case filter", err.Error())
		return
	}
	result, err := h.usecase.ListCompensationCases(c.Request.Context(), adminID, filter)
	if err != nil {
		writePaymentReadError(c, err, "Failed to list compensation cases")
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
	adminID, ok := requireAdmin(c)
	if !ok {
		return
	}
	caseID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid compensation case ID", err.Error())
		return
	}
	result, err := h.usecase.GetCompensationCase(c.Request.Context(), adminID, caseID)
	if err != nil {
		writePaymentReadError(c, err, "Failed to get compensation case")
		return
	}
	response.Success(c, "Compensation case retrieved successfully", result)
}

func requireAdmin(c *gin.Context) (uuid.UUID, bool) {
	return requireActor(c, "ADMIN", "Only administrators can view compensation cases")
}

func compensationFilterFromRequest(c *gin.Context) (apppayment.CompensationCaseFilter, error) {
	filter := apppayment.CompensationCaseFilter{Status: paymentdomain.CompensationStatus(strings.TrimSpace(c.Query("status"))), ReasonCode: paymentdomain.CompensationReasonCode(strings.TrimSpace(c.Query("reason_code")))}
	var err error
	if filter.ExpertID, err = optionalUUID(c.Query("expert_id")); err != nil {
		return filter, err
	}
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
