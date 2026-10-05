// cmd/main.go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	// Import các package nội bộ của dự án
	"booking-service/config"
	"booking-service/internal/infrastructure/client"
	"booking-service/internal/infrastructure/http/routes"

	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/application/schedule"
	"booking-service/internal/application/slot"
	"booking-service/internal/application/timeoff"
	appointmentrepo "booking-service/internal/infrastructure/persistence/repository/appointment"
	schedulerepo "booking-service/internal/infrastructure/persistence/repository/schedule"
	slotrepo "booking-service/internal/infrastructure/persistence/repository/slot"
	timeoffrepo "booking-service/internal/infrastructure/persistence/repository/timeoff"

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
	client.InitPublicKey(config.AppConfig.AuthServiceInternalURL)

	// 0.2 Khởi tạo TokenManager nội bộ — booking có thể gọi payment-service nội bộ
	_ = client.NewTokenManager(
		config.AppConfig.AuthServiceInternalURL,
		config.AppConfig.InternalClientID,
		config.AppConfig.InternalClientSecret,
	)
	log.Printf("🔐 Internal M2M auth initialized for client: %s", config.AppConfig.InternalClientID)

	// 1. Kết nối Database
	config.ConnectDB()

	// 2. Khởi tạo Router Gin
	router := gin.Default()

	// ---- KHỞI TẠO CÁC TẦNG LÕI ----

	// Repository
	scheduleRepo := schedulerepo.NewRepository(config.DB)
	slotRepo := slotrepo.NewRepository(config.DB)
	appointmentRepo := appointmentrepo.NewRepository(config.DB)
	timeoffRepo := timeoffrepo.NewRepository(config.DB)

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

	routes.RegisterSlotRoutes(publicAPI, privateAPI, slotRepo, appointmentRepo, slotUsecase, generationService)
	routes.RegisterAppointmentRoutes(publicAPI, privateAPI, internalAPI, appointmentUsecase)
	routes.RegisterTimeOffRoutes(privateAPI, timeoffUsecase)
	routes.RegisterScheduleRoutes(publicAPI, privateAPI, scheduleUsecase)

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
	appappointment.StartExpiredLockWorker(appointmentRepo)
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
