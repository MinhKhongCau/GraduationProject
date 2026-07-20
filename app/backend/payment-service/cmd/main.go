package main

import (
	"context"
	"log"
	"os"
	"time"

	"payment-service/internal/config"
	"payment-service/pkg/database"
	"payment-service/pkg/internal_auth"
	"payment-service/pkg/rabbitmq"
	"payment-service/pkg/redis"
	"payment-service/routes"

	paymentHTTP "payment-service/internal/payment/adapter/in/http"
	bookingrest "payment-service/internal/payment/adapter/out/bookingrest"
	"payment-service/internal/payment/adapter/out/outbox"
	paymentpostgres "payment-service/internal/payment/adapter/out/postgres"
	"payment-service/internal/payment/adapter/out/vnpay"
	apppayment "payment-service/internal/payment/application"

	"payment-service/internal/wallet"
	walletHandler "payment-service/internal/wallet/handler"

	"payment-service/internal/withdrawal"
	withdrawalHandler "payment-service/internal/withdrawal/handler"

	_ "payment-service/docs" // Import swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Payment Service API
// @version 1.0
// @description Hệ thống thanh toán và ví điện tử - MindCare
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// 1. Load Cấu hình & Biến môi trường
	config.LoadConfig()

	// 1.1 Cache RSA Public Key từ Auth Service (dùng để verify JWT user và internal JWT)
	internal_auth.InitPublicKey(config.AppConfig.AuthServiceInternalURL)

	// 1.2 Khởi tạo TokenManager nội bộ — dùng để GỌI sang service khác
	// TokenManager được lưu lại để inject vào BookingServiceClient
	tokenManager := internal_auth.NewTokenManager(
		config.AppConfig.AuthServiceInternalURL,
		config.AppConfig.InternalClientID,
		config.AppConfig.InternalClientSecret,
	)
	log.Printf("🔐 Internal M2M auth initialized for client: %s", config.AppConfig.InternalClientID)

	// 1.3 Khởi tạo BookingServiceClient (REST implementation)
	// Để chuyển sang gRPC sau này: chỉ đổi dòng này thành bookingClient.NewGrpcBookingClient(...)
	bookingSvcClient := bookingrest.NewRestBookingClient(
		config.AppConfig.BookingServiceInternalURL,
		tokenManager,
	)

	// 2. Kết nối CSDL & Chạy Migration
	database.ConnectDB()

	// 3. Kết nối hạ tầng RabbitMQ & Redis (hoạt động chế độ fallback nếu lỗi)
	rabbitmq.InitRabbitMQ()
	redis.InitRedis()

	// 4. Khởi tạo các tầng nghiệp vụ
	// Repositories
	walletRepo := wallet.NewRepository(database.DB)
	paymentRepo := paymentpostgres.NewRepository(database.DB)
	withdrawalRepo := withdrawal.NewRepository(database.DB)
	outboxRepo := outbox.NewRepository(database.DB)

	// Usecases
	walletUsecase := wallet.NewUsecase(walletRepo)

	vnpTmnCode := os.Getenv("VNP_TMN_CODE")
	vnpHashSecret := os.Getenv("VNP_HASH_SECRET")
	vnpPaymentURL := os.Getenv("VNP_PAYMENT_URL")
	vnpReturnURL := os.Getenv("VNP_RETURN_URL")
	vnpayClient := vnpay.NewVNPayClient(vnpTmnCode, vnpHashSecret, vnpPaymentURL, vnpReturnURL)

	paymentUoW := paymentpostgres.NewUnitOfWork(database.DB, walletUsecase)
	paymentUsecase := apppayment.NewUsecaseWithOptions(paymentRepo, paymentUoW, vnpayClient, bookingSvcClient, apppayment.Options{
		OrderTTL:      config.AppConfig.PaymentOrderTTL,
		MinimumWindow: config.AppConfig.PaymentMinUsableWindow,
	})
	withdrawalUsecase := withdrawal.NewUsecase(withdrawalRepo, walletUsecase)

	// Handlers
	wHandler := walletHandler.NewHandler(walletUsecase)
	pHandler := paymentHTTP.NewHandler(paymentUsecase)
	wdHandler := withdrawalHandler.NewHandler(withdrawalUsecase)

	// 5. Khởi chạy Tần số quét (Background Workers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Worker 1: Giải phóng tiền hold (Pending -> Available) sau 24h
	// Để thuận tiện test, mặc định hold 1 phút nếu không cấu hình env
	holdDuration := 24 * time.Hour
	if envHold := os.Getenv("HOLD_PERIOD_MINUTES"); envHold != "" {
		if min, err := time.ParseDuration(envHold + "m"); err == nil {
			holdDuration = min
		}
	} else {
		holdDuration = 1 * time.Minute // Mặc định dev là 1 phút
	}
	walletWorker := wallet.NewWorker(database.DB, walletRepo, holdDuration)
	go walletWorker.Start(ctx)

	// Worker 2: Xử lý timeout của các yêu cầu rút PROCESSING quá 5 phút
	withdrawalWorker := withdrawal.NewWorker(database.DB, withdrawalUsecase, 5*time.Minute)
	go withdrawalWorker.Start(ctx)

	// Worker 3: Quét Outbox events để dispatch (RabbitMQ / Internal REST Call sang Booking)
	outboxPublisher := outbox.NewPublisherWithOptions(outboxRepo, bookingSvcClient, outbox.PublisherOptions{
		PollInterval: config.AppConfig.OutboxPollInterval,
		MaxAttempts:  config.AppConfig.OutboxMaxAttempts,
		BaseBackoff:  config.AppConfig.OutboxBaseBackoff,
		MaxBackoff:   config.AppConfig.OutboxMaxBackoff,
		BatchSize:    config.AppConfig.OutboxBatchSize,
	})
	go outboxPublisher.Start(ctx)

	// Worker 4: Đối soát ví (Ledger Audit) & Đối soát giao dịch VNPay
	reconciliationWorker := wallet.NewReconciliationWorker(database.DB)
	go reconciliationWorker.Start(ctx)

	// 6. Khởi tạo Gin Router
	r := gin.Default()

	// Đăng ký API endpoints
	routes.SetupRoutes(r, wHandler, pHandler, wdHandler)

	// Swagger endpoint
	r.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Payment Service is running smoothly with clean architecture!",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "payment-service",
			"status":  "up and running",
			"db":      "connected",
		})
	})

	// 7. Chạy Server
	port := config.AppConfig.ServerPort
	log.Printf("🚀 Payment Service đang chạy tại http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Không thể khởi chạy Server: %v", err)
	}
}
