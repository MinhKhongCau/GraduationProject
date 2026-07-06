package schedules

import (
	"booking-service/internal/delivery/http/schedules/handler"
	"booking-service/internal/repository/postgres"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	publicGroup, privateGroup *gin.RouterGroup,
	generatorRepo *postgres.GeneratorRepository,
) {
	h := handler.NewScheduleHandler(generatorRepo)

	// 1. Templates (Công khai hoặc riêng tùy vai trò)
	publicGroup.GET("/templates", h.GetTemplates)
	privateGroup.POST("/templates", h.CreateTemplate) // Admin

	// 2. Availabilities (Chuyên gia)
	privateGroup.POST("/availabilities", h.CreateAvailability)
	privateGroup.GET("/availabilities", h.GetAvailabilities)
}
