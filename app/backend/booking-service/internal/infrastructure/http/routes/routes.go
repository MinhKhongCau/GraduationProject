package routes

import (
	appappointment "booking-service/internal/application/appointment"
	appschedule "booking-service/internal/application/schedule"
	appslot "booking-service/internal/application/slot"
	apptimeoff "booking-service/internal/application/timeoff"
	appointmenthandler "booking-service/internal/infrastructure/http/handlers/appointment"
	schedulehandler "booking-service/internal/infrastructure/http/handlers/schedule"
	slothandler "booking-service/internal/infrastructure/http/handlers/slot"
	timeoffhandler "booking-service/internal/infrastructure/http/handlers/timeoff"
	"booking-service/internal/infrastructure/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterSlotRoutes(
	publicGroup, privateGroup *gin.RouterGroup,
	repo appslot.Repository,
	appointmentRepo appappointment.Repository,
	usecase appslot.Usecase,
	generation appslot.GenerationApplication,
) {
	h := slothandler.NewHandler(repo, appointmentRepo, usecase, generation)

	// Public APIs
	publicGroup.GET("/slots/available-dates", h.GetDates)
	publicGroup.GET("/slots/available-times", h.GetTimes)

	// Private APIs
	privateGroup.POST("/slots/generate", h.Generate)
	privateGroup.POST("/slots/:id/lock", h.Lock)
	privateGroup.GET("/slots/expert", h.GetExpert)
}

// RegisterAppointmentRoutes đăng ký toàn bộ routes của appointment module.
// Nhận thêm router *gin.Engine để có thể mount /internal group ở root level.
func RegisterAppointmentRoutes(publicGroup, privateGroup *gin.RouterGroup, internalGroup *gin.RouterGroup, usecase appappointment.Usecase) {
	h := appointmenthandler.NewHandler(usecase)

	// ── Public routes (qua Kong Gateway, không cần user JWT) ────────────────
	_ = publicGroup // Public payment webhook intentionally disabled; use internal payment webhook only.

	// ── Private routes (yêu cầu JWT của User từ Kong) ──────────────────────
	privateGroup.POST("/appointments", h.Create)
	privateGroup.GET("/appointments", h.GetPatient)
	privateGroup.GET("/appointments/expert", h.GetExpert)
	privateGroup.GET("/appointments/:id", h.GetDetail)
	privateGroup.PATCH("/appointments/:id/cancel", h.Cancel)

	// ── Medical Record routes ──────────────────────────────────────────────
	privateGroup.POST("/appointments/:id/medical-record", h.SaveMedicalRecord)
	privateGroup.PUT("/appointments/:id/medical-record", h.SaveMedicalRecord)
	privateGroup.GET("/appointments/:id/medical-record", h.GetAppointmentMedicalRecord)
	privateGroup.GET("/medical-records", h.ListMedicalRecords)
	privateGroup.GET("/medical-records/:id", h.GetMedicalRecordDetail)

	// ── Internal routes (chỉ service nội bộ mới gọi được — M2M JWT) ─────────
	// Kong đã block /internal/* từ Internet → an toàn tuyệt đối.
	// Payment Service gọi vào đây sau khi nhận IPN thành công từ VNPay.
	internalAppt := internalGroup.Group("/appointments")
	internalAppt.Use(middleware.InternalAuth())
	{
		// POST /internal/appointments/:id/webhook
		internalAppt.POST("/:id/webhook", h.InternalPaymentWebhook)

		// GET /internal/appointments/:id
		internalAppt.GET("/:id", h.InternalGetAppointment)

		// POST /internal/appointments/:id/payment-eligibility
		internalAppt.POST("/:id/payment-eligibility", h.InternalPaymentEligibility)
	}
}

func RegisterTimeOffRoutes(
	privateGroup *gin.RouterGroup,
	usecase apptimeoff.Usecase,
) {
	h := timeoffhandler.NewHandler(usecase)

	timeoffGroup := privateGroup.Group("/time-off")
	{
		timeoffGroup.POST("", h.Create)
		timeoffGroup.POST("/confirm", h.Confirm)
		timeoffGroup.GET("", h.Get)
		timeoffGroup.DELETE("/:id", h.Delete)
	}
}

func RegisterScheduleRoutes(
	publicGroup, privateGroup *gin.RouterGroup,
	usecase appschedule.Usecase,
) {
	h := schedulehandler.NewHandler(usecase)

	// 1. Templates (Công khai hoặc riêng tùy vai trò)
	publicGroup.GET("/templates", h.GetTemplates)
	privateGroup.POST("/templates", h.CreateTemplate) // Admin
	privateGroup.GET("/templates", h.GetAdminTemplates)
	privateGroup.PATCH("/templates/:id", h.UpdateTemplate) // Admin

	// 2. Availabilities (Chuyên gia)
	privateGroup.POST("/availabilities", h.CreateAvailability)
	privateGroup.GET("/availabilities", h.GetAvailabilities)
	privateGroup.PATCH("/availabilities/:id", h.UpdateAvailability)
}
