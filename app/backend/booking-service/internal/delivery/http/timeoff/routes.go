package timeoff

import (
	"booking-service/internal/delivery/http/timeoff/handler"
	"booking-service/internal/repository/postgres"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	privateGroup *gin.RouterGroup,
	timeoffRepo *postgres.TimeOffRepository,
	slotRepo *postgres.SlotRepository,
	appointmentRepo *postgres.AppointmentRepository,
) {
	createHandler := handler.NewCreateTimeOffHandler(timeoffRepo, slotRepo, appointmentRepo)

	timeoffGroup := privateGroup.Group("/time-off")
	{
		timeoffGroup.POST("", createHandler.Handle)
	}
}
