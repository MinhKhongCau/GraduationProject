package main

import (
	"log"
	"os"

	"profile-service/config"
	"profile-service/internal/application/expertprofile"
	"profile-service/internal/infrastructure/http/handlers"
	"profile-service/internal/infrastructure/http/routes"
	"profile-service/internal/infrastructure/messaging"
	"profile-service/internal/infrastructure/messaging/consumer"
	"profile-service/internal/infrastructure/persistence/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "profile-service/docs" // Ignore error if it doesn't exist yet
)

// @title Profile Service API
// @version 1.0
// @description Hồ sơ người dùng (Admin/Patient/Expert) - MindCare
// @BasePath /
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

	// 2b. Lắng nghe sự kiện user.created từ auth-service (RabbitMQ)
	rabbitCfg := config.LoadRabbitMQConfig()
	messaging.StartUserCreatedConsumer(rabbitCfg, consumer.HandleUserCreated)

	// 3. Khởi tạo Gin Router
	r := gin.Default()

	// 3b. Composition root: Infrastructure -> Application -> HTTP adapter
	expertRepo := repository.NewExpertRepository(config.DB)
	specRepo := repository.NewSpecializationRepository(config.DB)
	eventPublisher := messaging.LogEventPublisher{}
	expertHandler := handlers.NewExpertHandler(
		expertprofile.NewReplaceExpertProfile(expertRepo, specRepo, eventPublisher),
		expertprofile.NewPatchExpertProfile(expertRepo, specRepo, eventPublisher),
	)

	routes.SetupRoutes(r, routes.Handlers{Expert: expertHandler})

	// Swagger endpoint
	r.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Profile Service is running smoothly!",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "profile-service",
			"status":  "up and running",
		})
	})

	// 4. Lấy port và chạy server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002" // Mặc định nếu quên setup
	}

	log.Printf("🚀 Server đang chạy tại http://localhost:%s", port)
	r.Run(":" + port)
}
