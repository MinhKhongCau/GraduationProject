// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	// Import các package nội bộ của dự án
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/database"

	"booking-service/internal/delivery/http/appointments"
	"booking-service/internal/delivery/http/slots"
	"booking-service/internal/worker"

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
	// 1. Kết nối Database
	database.ConnectDB()

	// 2. Khởi tạo Router Gin
	router := gin.Default()

	// ---- KHỞI TẠO CÁC TẦNG LÕI ----

	// Repository
	generatorRepo := postgres.NewGeneratorRepository(database.DB)
	slotRepo := postgres.NewSlotRepository(database.DB)
	appointmentRepo := postgres.NewAppointmentRepository(database.DB)

	// 3. Khai báo API Endpoints
	api := router.Group("/api/v1")
	slots.RegisterRoutes(api, slotRepo, generatorRepo, appointmentRepo)
	appointments.RegisterRoutes(api, appointmentRepo)

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

	// 5. Khởi chạy Background Worker: Dọn dẹp Expired Locks mỗi 60 giây
	worker.StartExpiredLockWorker(appointmentRepo)

	// 6. Khởi chạy Server
	log.Println("🚀 Starting Booking Service on port 8083...")
	if err := router.Run(":8083"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
