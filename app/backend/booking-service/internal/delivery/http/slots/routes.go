package slots

import (
	"booking-service/internal/delivery/http/slots/handler"
	"booking-service/internal/repository/postgres"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, slotRepo *postgres.SlotRepository, generatorRepo *postgres.GeneratorRepository, appointmentRepo *postgres.AppointmentRepository) {
	getDatesHandler := handler.NewGetAvailableDatesHandler(slotRepo)
	getTimesHandler := handler.NewGetAvailableTimesHandler(slotRepo)
	generateHandler := handler.NewGenerateSlotsHandler(generatorRepo)
	lockSlotHandler := handler.NewLockSlotHandler(appointmentRepo)

	slotsGroup := rg.Group("/slots")
	{
		slotsGroup.POST("/generate", generateHandler.Handle)
		slotsGroup.GET("/available-dates", getDatesHandler.Handle)
		slotsGroup.GET("/available-times", getTimesHandler.Handle)
		slotsGroup.POST("/:id/lock", lockSlotHandler.Handle)
	}
}
