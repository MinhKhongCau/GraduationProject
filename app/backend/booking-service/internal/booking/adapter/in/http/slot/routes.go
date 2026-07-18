package handler

import (
	"booking-service/internal/appointment"
	"booking-service/internal/slot"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	publicGroup, privateGroup *gin.RouterGroup,
	repo slot.Repository,
	appointmentRepo appointment.Repository,
	usecase slot.Usecase,
	scheduleRepo ScheduleProvider,
	timeoffRepo TimeOffProvider,
) {
	h := NewHandler(repo, appointmentRepo, usecase, scheduleRepo, timeoffRepo)

	// Public APIs
	publicGroup.GET("/slots/available-dates", h.GetDates)
	publicGroup.GET("/slots/available-times", h.GetTimes)

	// Private APIs
	privateGroup.POST("/slots/generate", h.Generate)
	privateGroup.POST("/slots/:id/lock", h.Lock)
	privateGroup.GET("/slots/expert", h.GetExpert)
}
