package handler

import (
	"booking-service/internal/timeoff"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	privateGroup *gin.RouterGroup,
	usecase timeoff.Usecase,
) {
	h := NewHandler(usecase)

	timeoffGroup := privateGroup.Group("/time-off")
	{
		timeoffGroup.POST("", h.Create)
		timeoffGroup.POST("/confirm", h.Confirm)
		timeoffGroup.GET("", h.Get)
		timeoffGroup.DELETE("/:id", h.Delete)
	}
}
