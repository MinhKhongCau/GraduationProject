package slots

import (
	"booking-service/internal/delivery/http/slots/handler"
	"booking-service/internal/repository/postgres"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(publicGroup, privateGroup *gin.RouterGroup, slotRepo *postgres.SlotRepository, generatorRepo *postgres.GeneratorRepository, appointmentRepo *postgres.AppointmentRepository) {
	getDatesHandler := handler.NewGetAvailableDatesHandler(slotRepo)
	getTimesHandler := handler.NewGetAvailableTimesHandler(slotRepo)
	generateHandler := handler.NewGenerateSlotsHandler(generatorRepo)
	lockSlotHandler := handler.NewLockSlotHandler(appointmentRepo)

	publicSlotsGroup := publicGroup.Group("/slots")
	{
		publicSlotsGroup.GET("/available-dates", getDatesHandler.Handle)
		publicSlotsGroup.GET("/available-times", getTimesHandler.Handle)
	}

	privateSlotsGroup := privateGroup.Group("/slots")
	{
		privateSlotsGroup.POST("/generate", generateHandler.Handle)
		privateSlotsGroup.POST("/:id/lock", lockSlotHandler.Handle)
	}
}
