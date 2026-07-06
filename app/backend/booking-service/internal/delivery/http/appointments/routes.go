package appointments

import (
	"booking-service/internal/delivery/http/appointments/handler"
	"booking-service/internal/repository/postgres"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(publicGroup, privateGroup *gin.RouterGroup, appointmentRepo *postgres.AppointmentRepository) {
	createHandler := handler.NewCreateAppointmentHandler(appointmentRepo)
	webhookHandler := handler.NewPaymentWebhookHandler(appointmentRepo)

	publicAppointmentsGroup := publicGroup.Group("/appointments")
	{
		publicAppointmentsGroup.POST("/webhook", webhookHandler.Handle)
	}

	privateAppointmentsGroup := privateGroup.Group("/appointments")
	{
		privateAppointmentsGroup.POST("", createHandler.Handle)
	}
}
