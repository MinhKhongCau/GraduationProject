package payment

import (
	"net/http"
	apppayment "payment-service/internal/application/payment"
	paymentdomain "payment-service/internal/domain/payment"
	"payment-service/internal/infrastructure/http/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type reviewOrderRequest struct {
	// MANUAL_REVIEW hoặc REFUND_REQUIRED
	Action string `json:"action" binding:"required"`
	Note   string `json:"note"`
}

type resolveCompensationRequest struct {
	Note string `json:"note"`
}

// ---------- EXPERT ----------

// ListExpertOrders handles GET /api/v1/payments/expert/orders.
// @Summary [EXPERT] List payment orders paid to me (gross, commission, net)
// @Tags Expert - Payments
// @Security BearerAuth
// @Param status query string false "PENDING, SUCCESS, FAILED, EXPIRED"
// @Param type query string false "APPOINTMENT or TOP_UP"
// @Param fulfillment_status query string false "Fulfillment status"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /payments/expert/orders [get]
func (h *Handler) ListExpertOrders(c *gin.Context) {
	expertID, filter, ok := h.expertRequest(c)
	if !ok {
		return
	}
	result, err := h.manager.ListExpertOrders(c.Request.Context(), expertID, filter)
	if err != nil {
		writePaymentReadError(c, err, "Failed to list payment orders")
		return
	}
	response.Success(c, "Payment orders retrieved successfully", result)
}

// SummarizeExpertOrders handles GET /api/v1/payments/expert/orders/summary.
// @Summary [EXPERT] Summarize my revenue: gross, commission and net amount
// @Tags Expert - Payments
// @Security BearerAuth
// @Param status query string false "PENDING, SUCCESS, FAILED, EXPIRED"
// @Param type query string false "APPOINTMENT or TOP_UP"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Router /payments/expert/orders/summary [get]
func (h *Handler) SummarizeExpertOrders(c *gin.Context) {
	expertID, filter, ok := h.expertRequest(c)
	if !ok {
		return
	}
	result, err := h.manager.SummarizeExpertOrders(c.Request.Context(), expertID, filter)
	if err != nil {
		writePaymentReadError(c, err, "Failed to summarize payment orders")
		return
	}
	response.Success(c, "Payment order summary retrieved successfully", result)
}

// GetExpertOrder handles GET /api/v1/payments/expert/orders/:id.
// @Summary [EXPERT] Get a payment order paid to me
// @Tags Expert - Payments
// @Security BearerAuth
// @Param id path string true "Payment order UUID"
// @Router /payments/expert/orders/{id} [get]
func (h *Handler) GetExpertOrder(c *gin.Context) {
	expertID, ok := requireActor(c, "EXPERT", "Only experts can view their payment orders")
	if !ok || !h.requireManager(c) {
		return
	}
	orderID, ok := pathOrderID(c)
	if !ok {
		return
	}
	result, err := h.manager.GetExpertOrder(c.Request.Context(), expertID, orderID)
	if err != nil {
		writePaymentReadError(c, err, "Failed to get payment order")
		return
	}
	response.Success(c, "Payment order retrieved successfully", result)
}

// ---------- ADMIN (chỉ chuyên gia mình đã duyệt) ----------

// ListAdminOrders handles GET /api/v1/payments/admin/orders.
// @Summary [ADMIN] List payment orders of experts I manage
// @Tags Admin - Payments
// @Security BearerAuth
// @Param expert_id query string false "Expert auth UUID (must be managed by me)"
// @Param payer_id query string false "Patient auth UUID"
// @Param appointment_id query string false "Appointment UUID"
// @Param status query string false "PENDING, SUCCESS, FAILED, EXPIRED"
// @Param type query string false "APPOINTMENT or TOP_UP"
// @Param fulfillment_status query string false "Fulfillment status"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Zero-based page"
// @Param size query int false "Page size, 1-100"
// @Router /payments/admin/orders [get]
func (h *Handler) ListAdminOrders(c *gin.Context) {
	adminID, filter, ok := h.adminRequest(c)
	if !ok {
		return
	}
	result, err := h.manager.ListAdminOrders(c.Request.Context(), adminID, filter)
	if err != nil {
		writePaymentReadError(c, err, "Failed to list payment orders")
		return
	}
	response.Success(c, "Payment orders retrieved successfully", result)
}

// SummarizeAdminOrders handles GET /api/v1/payments/admin/orders/summary.
// @Summary [ADMIN] Summarize payment orders of experts I manage
// @Tags Admin - Payments
// @Security BearerAuth
// @Param expert_id query string false "Expert auth UUID (must be managed by me)"
// @Param payer_id query string false "Patient auth UUID"
// @Param status query string false "PENDING, SUCCESS, FAILED, EXPIRED"
// @Param type query string false "APPOINTMENT or TOP_UP"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Router /payments/admin/orders/summary [get]
func (h *Handler) SummarizeAdminOrders(c *gin.Context) {
	adminID, filter, ok := h.adminRequest(c)
	if !ok {
		return
	}
	result, err := h.manager.SummarizeAdminOrders(c.Request.Context(), adminID, filter)
	if err != nil {
		writePaymentReadError(c, err, "Failed to summarize payment orders")
		return
	}
	response.Success(c, "Payment order summary retrieved successfully", result)
}

// GetAdminOrder handles GET /api/v1/payments/admin/orders/:id.
// @Summary [ADMIN] Get full payment order detail (expert must be managed by me)
// @Tags Admin - Payments
// @Security BearerAuth
// @Param id path string true "Payment order UUID"
// @Router /payments/admin/orders/{id} [get]
func (h *Handler) GetAdminOrder(c *gin.Context) {
	adminID, ok := requireActor(c, "ADMIN", "Only administrators can manage payment orders")
	if !ok || !h.requireManager(c) {
		return
	}
	orderID, ok := pathOrderID(c)
	if !ok {
		return
	}
	result, err := h.manager.GetAdminOrder(c.Request.Context(), adminID, orderID)
	if err != nil {
		writePaymentReadError(c, err, "Failed to get payment order")
		return
	}
	response.Success(c, "Payment order retrieved successfully", result)
}

// ReviewOrder handles POST /api/v1/payments/admin/orders/:id/review.
// @Summary [ADMIN] Flag a paid order for manual review or refund
// @Tags Admin - Payments
// @Security BearerAuth
// @Accept json
// @Param id path string true "Payment order UUID"
// @Param request body reviewOrderRequest true "action: MANUAL_REVIEW or REFUND_REQUIRED"
// @Router /payments/admin/orders/{id}/review [post]
func (h *Handler) ReviewOrder(c *gin.Context) {
	adminID, ok := requireActor(c, "ADMIN", "Only administrators can manage payment orders")
	if !ok || !h.requireManager(c) {
		return
	}
	orderID, ok := pathOrderID(c)
	if !ok {
		return
	}
	var req reviewOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
	action := paymentdomain.CompensationStatus(strings.ToUpper(strings.TrimSpace(req.Action)))
	result, err := h.manager.ReviewOrder(c.Request.Context(), adminID, orderID, action, req.Note)
	if err != nil {
		writePaymentReadError(c, err, "Failed to review payment order")
		return
	}
	response.Success(c, "Payment order sent to review successfully", result)
}

// ResolveCompensationCase handles POST /api/v1/payments/compensation-cases/:id/resolve.
// @Summary [ADMIN] Resolve a compensation case of an expert I manage
// @Tags Admin - Payments
// @Security BearerAuth
// @Accept json
// @Param id path string true "Compensation case UUID"
// @Param request body resolveCompensationRequest false "Resolution note"
// @Router /payments/compensation-cases/{id}/resolve [post]
func (h *Handler) ResolveCompensationCase(c *gin.Context) {
	adminID, ok := requireActor(c, "ADMIN", "Only administrators can resolve compensation cases")
	if !ok || !h.requireManager(c) {
		return
	}
	caseID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid compensation case ID", err.Error())
		return
	}
	var req resolveCompensationRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
			return
		}
	}
	result, err := h.manager.ResolveCompensationCase(c.Request.Context(), adminID, caseID, req.Note)
	if err != nil {
		writePaymentReadError(c, err, "Failed to resolve compensation case")
		return
	}
	response.Success(c, "Compensation case resolved successfully", result)
}

func (h *Handler) expertRequest(c *gin.Context) (uuid.UUID, apppayment.PaymentOrderFilter, bool) {
	expertID, ok := requireActor(c, "EXPERT", "Only experts can view their payment orders")
	if !ok || !h.requireManager(c) {
		return uuid.Nil, apppayment.PaymentOrderFilter{}, false
	}
	filter, err := paymentOrderFilter(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment order filter", err.Error())
		return uuid.Nil, filter, false
	}
	return expertID, filter, true
}

func (h *Handler) adminRequest(c *gin.Context) (uuid.UUID, apppayment.PaymentOrderFilter, bool) {
	adminID, ok := requireActor(c, "ADMIN", "Only administrators can manage payment orders")
	if !ok || !h.requireManager(c) {
		return uuid.Nil, apppayment.PaymentOrderFilter{}, false
	}
	filter, err := paymentOrderFilter(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payment order filter", err.Error())
		return uuid.Nil, filter, false
	}
	for param, target := range map[string]*uuid.UUID{"expert_id": &filter.ExpertID, "payer_id": &filter.PayerID} {
		value, err := optionalUUID(c.Query(param))
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid payment order filter", param+": "+err.Error())
			return uuid.Nil, filter, false
		}
		if value != nil {
			*target = *value
		}
	}
	return adminID, filter, true
}

func (h *Handler) requireManager(c *gin.Context) bool {
	if h.manager == nil {
		response.Error(c, http.StatusInternalServerError, "Payment order management unavailable", "payment order manager unavailable")
		return false
	}
	return true
}
