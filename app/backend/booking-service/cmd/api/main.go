// cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	// Import các package nội bộ của dự án
	"booking-service/internal/config"
	"booking-service/pkg/database"
	"booking-service/pkg/internal_auth"

	"booking-service/internal/appointment"
	apptHandler "booking-service/internal/appointment/handler"
	"booking-service/internal/schedule"
	schedHandler "booking-service/internal/schedule/handler"
	"booking-service/internal/slot"
	slotHandler "booking-service/internal/slot/handler"
	"booking-service/internal/timeoff"
	timeoffHandler "booking-service/internal/timeoff/handler"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "booking-service/docs" // Ignore error if it doesn't exist yet
)

// @title Booking Service API
// @version 1.0
// @description Hệ thống đặt lịch khám bệnh - MindCare
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// 0. Load Configuration
	config.LoadConfig()

	// 0.1 Cache RSA Public Key từ Auth Service (verify cả JWT user lẫn internal JWT)
	internal_auth.InitPublicKey(config.AppConfig.AuthServiceInternalURL)

	// 0.2 Khởi tạo TokenManager nội bộ — booking có thể gọi payment-service nội bộ
	_ = internal_auth.NewTokenManager(
		config.AppConfig.AuthServiceInternalURL,
		config.AppConfig.InternalClientID,
		config.AppConfig.InternalClientSecret,
	)
	log.Printf("🔐 Internal M2M auth initialized for client: %s", config.AppConfig.InternalClientID)

	// 1. Kết nối Database
	database.ConnectDB()

	// 2. Khởi tạo Router Gin
	router := gin.Default()

	// ---- KHỞI TẠO CÁC TẦNG LÕI ----

	// Repository
	scheduleRepo := schedule.NewRepository(database.DB)
	slotRepo := slot.NewRepository(database.DB)
	appointmentRepo := appointment.NewRepository(database.DB)
	timeoffRepo := timeoff.NewRepository(database.DB)
	
	// Usecase
	slotUsecase := slot.NewUsecase(slotRepo, appointmentRepo)
	scheduleUsecase := schedule.NewUsecase(scheduleRepo)
	timeoffUsecase := timeoff.NewUsecase(timeoffRepo, slotRepo, appointmentRepo)
	appointmentUsecase := appointment.NewUsecase(appointmentRepo)

	// 3. Khai báo API Endpoints
	publicAPI := router.Group("/api/v1/public/booking")
	privateAPI := router.Group("/api/v1/booking")
	// internalAPI: chỉ service khác trong Docker network gọi được (Kong đã block /internal/* từ internet)
	internalAPI := router.Group("/internal")

	slotHandler.RegisterRoutes(publicAPI, privateAPI, slotRepo, appointmentRepo, slotUsecase, scheduleRepo, timeoffRepo)
	apptHandler.RegisterRoutes(publicAPI, privateAPI, internalAPI, appointmentUsecase)
	timeoffHandler.RegisterRoutes(privateAPI, timeoffUsecase)
	schedHandler.RegisterRoutes(publicAPI, privateAPI, scheduleUsecase)

	// 3.5. Swagger endpoint
	router.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 4. Health check API
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "booking-service",
			"status":  "up and running",
			"db":      "connected",
		})
	})

	// 5. Khởi chạy Background Workers
	slot.StartExpiredLockWorker(appointmentRepo)
	timeoff.StartWorker(timeoffUsecase)

	// 6. Khởi chạy Server
	port := config.AppConfig.ServerPort
	log.Printf("🚀 Starting Booking Service on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
