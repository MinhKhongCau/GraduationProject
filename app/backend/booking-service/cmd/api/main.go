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

	httpDelivery "booking-service/internal/delivery/http"

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

	// Handler
	generatorHandler := httpDelivery.NewGeneratorHandler(generatorRepo)
	slotHandler := httpDelivery.NewSlotHandler(slotRepo)
	appointmentHandler := httpDelivery.NewAppointmentHandler(appointmentRepo)

	// 3. Khai báo API Endpoints
	api := router.Group("/api/v1")
	{
		// === GIAI ĐOẠN 1: Expert Sinh Lịch ===
		// Yêu cầu Header từ Gateway: X-User-Role=EXPERT, X-User-Id=<expert_uuid>
		api.POST("/slots/generate", generatorHandler.HandleGenerateSlots)

		// === GIAI ĐOẠN 2: Patient Xem Lịch Trống ===
		api.GET("/slots/available-dates", slotHandler.HandleGetAvailableDates)   // ?expert_id=xxx
		api.GET("/slots/available-times", slotHandler.HandleGetAvailableTimes)   // ?date=YYYY-MM-DD&expert_id=xxx

		// === GIAI ĐOẠN 3: Khóa Chỗ, Đặt Lịch & Thanh Toán ===
		// Yêu cầu Header từ Gateway: X-User-Role=PATIENT, X-User-Id=<patient_uuid>
		api.POST("/slots/:id/lock", appointmentHandler.HandleLockSlot)
		api.POST("/appointments", appointmentHandler.HandleCreateAppointment)
		api.POST("/appointments/webhook", appointmentHandler.HandlePaymentWebhook)
	}

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
	// Worker này thay thế cho việc dùng Cronjob bên ngoài, chạy ngầm trong cùng process
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		log.Println("🔄 Expired Lock Worker đã khởi động, quét mỗi 60 giây...")

		for range ticker.C {
			cleaned, err := appointmentRepo.CancelExpiredLocks()
			if err != nil {
				log.Printf("⚠️  Worker lỗi khi dọn expired locks: %v", err)
			} else if cleaned > 0 {
				log.Printf("🧹 Worker đã dọn %d slot hết hạn, mở lại cho bệnh nhân khác.", cleaned)
			}
		}
	}()

	// 6. Khởi chạy Server
	log.Println("🚀 Starting Booking Service on port 8083...")
	if err := router.Run(":8083"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
