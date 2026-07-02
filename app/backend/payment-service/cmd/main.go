package main

import (
	"log"
	"os"

	"payment-service/config"
	"payment-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Payment Service is running smoothly!",
		})
	})

	// 4. Lấy port và chạy server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003" // Mặc định nếu quên setup
	}

	log.Printf("🚀 Server đang chạy tại http://localhost:%s", port)
	r.Run(":" + port)
}
