package handler

import (
	"booking-service/internal/appointment"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(publicGroup, privateGroup *gin.RouterGroup, usecase appointment.Usecase) {
	h := NewHandler(usecase)

	publicGroup.POST("/appointments/webhook", h.Webhook)

	privateGroup.POST("/appointments", h.Create)
	privateGroup.GET("/appointments", h.GetPatient)
	privateGroup.GET("/appointments/expert", h.GetExpert)
	privateGroup.PATCH("/appointments/:id/cancel", h.Cancel)
}
