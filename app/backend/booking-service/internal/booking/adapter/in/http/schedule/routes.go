package handler

import (
	"booking-service/internal/schedule"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	publicGroup, privateGroup *gin.RouterGroup,
	usecase schedule.Usecase,
) {
	h := NewHandler(usecase)

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
