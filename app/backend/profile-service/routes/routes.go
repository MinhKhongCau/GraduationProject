// File: routes/routes.go
package routes

import (
	"profile-service/internal/handlers"
	"profile-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// ---------- Internal API (service-to-service, KHÔNG đi qua Gateway) ----------
	// auth-service gọi thẳng vào đây (network nội bộ) ngay sau khi tạo tài khoản mới.
	internal := r.Group("/internal/api/v1/profiles")
	{
		internal.POST("/create", handlers.CreateProfileInternal)
		internal.POST("/sync-seed-authors", handlers.SyncSeedAuthorsInternal)
	}

	// ---------- Public API (đi qua Gateway) ----------
	api := r.Group("/api/v1/profiles")
	{
		// --- Self-service: /me ---
		self := api.Group("/me")
		self.Use(middleware.RequireAuth())
		{
			self.GET("", handlers.GetMe)
			self.PUT("", handlers.UpdateMe)
			self.PATCH("", handlers.PatchMe)
			self.GET("/medical-histories", handlers.ListMyMedicalHistories)
			self.POST("/medical-histories", handlers.AddMyMedicalHistory)
		}

		// --- Duyệt danh sách chuyên gia/chuyên khoa: public ---
		api.GET("/experts", handlers.ListExperts)
		api.GET("/experts/:id", handlers.GetExpert)
		api.GET("/specializations", handlers.GetAllSpecializations)
		api.GET("/patients/:id/public", middleware.RequireAuth(), handlers.GetPatientPublic)
		api.GET("/public/:id", handlers.GetPublicProfile)

		// --- Quản trị: chỉ ADMIN ---
		admin := api.Group("")
		admin.Use(middleware.RequireAuth(), middleware.RequireRole("ADMIN"))
		{
			// Quản lý chung mọi profile
			admin.GET("", handlers.ListProfiles)
			admin.POST("", handlers.CreateProfile)
			admin.GET("/:id", handlers.GetProfile)
			admin.PUT("/:id", handlers.UpdateProfile)
			admin.PATCH("/:id", handlers.PatchProfile)
			admin.DELETE("/:id", handlers.DeleteProfile)

			// Bệnh nhân
			admin.GET("/patients", handlers.ListPatients)
			admin.GET("/patients/:id", handlers.GetPatient)
			admin.PUT("/patients/:id", handlers.UpdatePatient)
			admin.PATCH("/patients/:id", handlers.PatchPatient)
			admin.GET("/patients/:id/medical-histories", handlers.ListPatientMedicalHistories)

			// Chuyên gia (duyệt hồ sơ, chỉnh sửa thay mặt)
			admin.PUT("/experts/:id", handlers.UpdateExpert)
			admin.PATCH("/experts/:id", handlers.PatchExpert)

			// Chuyên khoa
			admin.POST("/specializations", handlers.CreateSpecialization)
		}
	}
}
