// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	// Import các package nội bộ của dự án
	"booking-service/internal/config"
	"booking-service/pkg/database"
	"booking-service/pkg/internal_auth"

	apptHandler "booking-service/internal/booking/adapter/in/http/appointment"
	schedHandler "booking-service/internal/booking/adapter/in/http/schedule"
	slotHandler "booking-service/internal/booking/adapter/in/http/slot"
	timeoffHandler "booking-service/internal/booking/adapter/in/http/timeoff"
	appointmentpostgres "booking-service/internal/booking/adapter/out/postgres/appointment"
	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/schedule"
	"booking-service/internal/slot"
	"booking-service/internal/timeoff"

	_ "booking-service/docs" // Ignore error if it doesn't exist yet
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	appointmentRepo := appointmentpostgres.NewRepository(database.DB)
	timeoffRepo := timeoff.NewRepository(database.DB)

	// Usecase
	slotUsecase := slot.NewUsecase(slotRepo, appointmentRepo)
	scheduleUsecase := schedule.NewUsecaseWithReconciliation(scheduleRepo, timeoffRepo, config.AppConfig.RollingSlotDays)
	timeoffUsecase := timeoff.NewUsecase(timeoffRepo, slotRepo, appointmentRepo)
	appointmentUsecase := appappointment.NewUsecase(appointmentRepo)
	generationService := slot.NewGenerationService(slotUsecase, scheduleRepo, timeoffRepo)

	// 3. Khai báo API Endpoints
	publicAPI := router.Group("/api/v1/public/booking")
	privateAPI := router.Group("/api/v1/booking")
	// internalAPI: chỉ service khác trong Docker network gọi được (Kong đã block /internal/* từ internet)
	internalAPI := router.Group("/internal")

	slotHandler.RegisterRoutes(publicAPI, privateAPI, slotRepo, appointmentRepo, slotUsecase, generationService)
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
	workerContext := context.Background()
	slot.StartExpiredLockWorker(appointmentRepo)
	timeoff.StartWorkerWithContext(workerContext, timeoffUsecase, timeoff.WorkerInterval)
	go slot.RunStartupGeneration(workerContext, generationService, config.AppConfig.RollingSlotDays)
	slot.StartRollingGenerationWorker(workerContext, generationService, config.AppConfig.RollingSlotDays)

	// 6. Khởi chạy Server
	port := config.AppConfig.ServerPort
	log.Printf("🚀 Starting Booking Service on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
