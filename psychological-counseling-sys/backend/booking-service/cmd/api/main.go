// cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	// Import các package nội bộ của dự án
	"booking-service/internal/repository/postgres"
	"booking-service/pkg/database"

	httpDelivery "booking-service/internal/delivery/http"
)

func main() {
	// 1. Kết nối Database
	database.ConnectDB()

	// 2. Khởi tạo Router Gin
	router := gin.Default()

	// ---- KHU VỰC KHỞI TẠO CÁC TẦNG LÕI ----
	generatorRepo := postgres.NewGeneratorRepository(database.DB)

	// Khởi tạo Handler (Sử dụng bí danh httpDelivery)
	generatorHandler := httpDelivery.NewGeneratorHandler(generatorRepo)

	slotRepo := postgres.NewSlotRepository(database.DB)
	slotHandler := httpDelivery.NewSlotHandler(slotRepo)

	// 3. Khai báo API Endpoint mới
	api := router.Group("/api/v1")
	{
		// Kích hoạt API sinh lịch tự động
		api.POST("/slots/generate", generatorHandler.HandleGenerateSlots)
		api.GET("/slots/available-dates", slotHandler.HandleGetAvailableDates)
		api.GET("/slots/available-times", slotHandler.HandleGetAvailableTimes)
	}

	// 4. Health check API
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "booking-service",
			"status":  "up and running",
			"db":      "connected",
		})
	})

	// 5. Khởi chạy Server
	log.Println("Starting Booking Service on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
