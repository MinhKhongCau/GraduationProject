package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleVNPayIPN handles GET /api/v1/payments/vnpay-ipn
// @Summary      [PUBLIC/WEBHOOK] Receive VNPay payment webhook (IPN)
// @Description  VNPay calls this API to update the payment status of an order. Role: Public (no token required, verifies checksum).
// @Tags         Payment Orders
// @Produce      json
// @Param        queryParams  query     object  false  "VNPay automated response parameters"
// @Success      200          {object}  map[string]string
// @Router       /payments/vnpay-ipn [get]
func (h *Handler) HandleVNPayIPN(c *gin.Context) {
	queryParams := c.Request.URL.Query()

	alreadyProcessed, err := h.usecase.ProcessIPN(c.Request.Context(), queryParams)
	if err != nil {
		logStr := err.Error()
		classification := strings.ToLower(logStr)
		if strings.Contains(classification, "checksum") || strings.Contains(classification, "securehash") {
			c.JSON(http.StatusOK, gin.H{"RspCode": "97", "Message": "Invalid Signature"})
			return
		}
		if strings.Contains(classification, "not found") {
			c.JSON(http.StatusOK, gin.H{"RspCode": "01", "Message": "Order not found"})
			return
		}
		if strings.Contains(classification, "amount") {
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
