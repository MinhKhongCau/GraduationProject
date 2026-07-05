package appointments

import (
	"booking-service/internal/delivery/http/appointments/handler"
	"booking-service/internal/repository/postgres"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, appointmentRepo *postgres.AppointmentRepository) {
	createHandler := handler.NewCreateAppointmentHandler(appointmentRepo)
	webhookHandler := handler.NewPaymentWebhookHandler(appointmentRepo)

	appointmentsGroup := rg.Group("/appointments")
	{
		appointmentsGroup.POST("", createHandler.Handle)
		appointmentsGroup.POST("/webhook", webhookHandler.Handle)
	}
}
