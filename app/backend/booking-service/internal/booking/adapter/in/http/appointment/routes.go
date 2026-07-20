package handler

import (
	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/pkg/internal_auth"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký toàn bộ routes của appointment module.
// Nhận thêm router *gin.Engine để có thể mount /internal group ở root level.
func RegisterRoutes(publicGroup, privateGroup *gin.RouterGroup, internalGroup *gin.RouterGroup, usecase appappointment.Usecase) {
	h := NewHandler(usecase)

	// ── Public routes (qua Kong Gateway, không cần user JWT) ────────────────
	_ = publicGroup // Public payment webhook intentionally disabled; use internal payment webhook only.

	// ── Private routes (yêu cầu JWT của User từ Kong) ──────────────────────
	privateGroup.POST("/appointments", h.Create)
	privateGroup.GET("/appointments", h.GetPatient)
	privateGroup.GET("/appointments/expert", h.GetExpert)
	privateGroup.GET("/appointments/:id", h.GetDetail)
	privateGroup.PATCH("/appointments/:id/cancel", h.Cancel)

	// ── Internal routes (chỉ service nội bộ mới gọi được — M2M JWT) ─────────
	// Kong đã block /internal/* từ Internet → an toàn tuyệt đối.
	// Payment Service gọi vào đây sau khi nhận IPN thành công từ VNPay.
	internalAppt := internalGroup.Group("/appointments")
	internalAppt.Use(internal_auth.Middleware())
	{
		// POST /internal/appointments/:id/webhook
		internalAppt.POST("/:id/webhook", h.InternalPaymentWebhook)

		// GET /internal/appointments/:id
		internalAppt.GET("/:id", h.InternalGetAppointment)

		// POST /internal/appointments/:id/payment-eligibility
		internalAppt.POST("/:id/payment-eligibility", h.InternalPaymentEligibility)
	}
}
