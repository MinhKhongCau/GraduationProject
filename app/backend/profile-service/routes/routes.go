// File: routes/routes.go
package routes

import (
	"profile-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/profiles")
	{
		// Nhóm API dành cho Patient
		patients := api.Group("/patients")
		{
			// Truyền account_id tạm qua URL để test (Sau này sẽ đổi thành quét từ JWT)
			patients.GET("/:account_id", handlers.GetPatientProfile)
			patients.POST("/:account_id", handlers.CreatePatientProfile)

			patients.GET("/:account_id/medical-histories", handlers.GetMedicalHistories)
			patients.POST("/:account_id/medical-histories", handlers.AddMedicalHistory)
		}

		specs := api.Group("/specializations")
		{
			specs.GET("/", handlers.GetAllSpecializations)
			specs.POST("/", handlers.CreateSpecialization)
		}

		experts := api.Group("/experts")
		{
			// Truyền account_id của Bác sĩ trên URL
			experts.POST("/:account_id", handlers.CreateExpertProfile)
			experts.GET("/", handlers.GetAllExperts)
			experts.GET("/:account_id", handlers.GetExpertProfile)
		}
	}
}
