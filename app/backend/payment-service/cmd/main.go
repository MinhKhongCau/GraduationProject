package main

import (
	"log"
	"os"

	"payment-service/config"
	"payment-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "payment-service/docs" // Ignore error if it doesn't exist yet
)

// @title Payment Service API
// @version 1.0
// @description Hệ thống thanh toán và ví điện tử - MindCare
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// 1. Load biến môi trường từ file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Cảnh báo: Không tìm thấy file .env, sẽ dùng biến môi trường của hệ thống")
	}

	// 2. Kết nối Database & Chạy Migration
	config.ConnectDB()

	// 3. Khởi tạo Gin Router
	r := gin.Default()

	routes.SetupRoutes(r)

	// Swagger endpoint
	r.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Payment Service is running smoothly!",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "payment-service",
			"status":  "up and running",
		})
	})

	// 4. Lấy port và chạy server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Mặc định nếu quên setup
	}

	log.Printf("🚀 Server đang chạy tại http://localhost:%s", port)
	r.Run(":" + port)
}
