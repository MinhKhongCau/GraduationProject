package main

import (
	"log"
	"os"

	"profile-service/config"
	"profile-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "profile-service/docs" // Ignore error if it doesn't exist yet
)

// @title Profile Service API
// @version 1.0
// @description Hệ thống quản lý hồ sơ người dùng - MindCare
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
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Profile Service is running smoothly!",
		})
	})

	// 4. Lấy port và chạy server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Mặc định nếu quên setup
	}

	log.Printf("🚀 Server đang chạy tại http://localhost:%s", port)
	r.Run(":" + port)
}
