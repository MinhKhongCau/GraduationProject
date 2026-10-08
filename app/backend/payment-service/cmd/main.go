package main

import (
	"context"
	"log"
	"os"
	"time"

	"payment-service/config"
	apppayment "payment-service/internal/application/payment"
	appwallet "payment-service/internal/application/wallet"
	appwithdrawal "payment-service/internal/application/withdrawal"
	"payment-service/internal/infrastructure/client/bookinggrpc"
	"payment-service/internal/infrastructure/client/bookingrest"
	"payment-service/internal/infrastructure/client/internalauth"
	"payment-service/internal/infrastructure/client/profilegrpc"
	"payment-service/internal/infrastructure/client/vnpay"
	paymentHandler "payment-service/internal/infrastructure/http/handlers/payment"
	walletHandler "payment-service/internal/infrastructure/http/handlers/wallet"
	withdrawalHandler "payment-service/internal/infrastructure/http/handlers/withdrawal"
	"payment-service/internal/infrastructure/http/routes"
	"payment-service/internal/infrastructure/messaging"
	"payment-service/internal/infrastructure/persistence/repository"

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
	internalauth.InitPublicKey(config.AppConfig.AuthServiceInternalURL)

	// 1.2 Khởi tạo TokenManager nội bộ — dùng để GỌI sang service khác
	// TokenManager được lưu lại để inject vào BookingServiceClient
	tokenManager := internalauth.NewTokenManager(
		config.AppConfig.AuthServiceInternalURL,
		config.AppConfig.InternalClientID,
		config.AppConfig.InternalClientSecret,
	)
	log.Printf("🔐 Internal M2M auth initialized for client: %s", config.AppConfig.InternalClientID)

	// 1.3 Khởi tạo BookingServiceClient: kết quả thanh toán + trạng thái lịch hẹn đi qua gRPC,
	// kiểm tra điều kiện thanh toán (eligibility) vẫn dùng REST.
	bookingRestClient := bookingrest.NewRestBookingClient(
		config.AppConfig.BookingServiceInternalURL,
		tokenManager,
	)
	bookingSvcClient, err := bookinggrpc.New(config.AppConfig.BookingGRPCAddr, tokenManager, bookingRestClient)
	if err != nil {
		log.Fatalf("Failed to init booking gRPC client: %v", err)
	}
	defer bookingSvcClient.Close()
	log.Printf("🛰️  Booking gRPC client target: %s", config.AppConfig.BookingGRPCAddr)

	// 1.4 profile-service gRPC: danh sách chuyên gia mà Admin quản lý (Admin đã duyệt chuyên gia)
	profileClient, err := profilegrpc.New(config.AppConfig.ProfileGRPCAddr)
	if err != nil {
		log.Fatalf("Failed to init profile gRPC client: %v", err)
	}
	defer profileClient.Close()
	log.Printf("🛰️  Profile gRPC client target: %s", config.AppConfig.ProfileGRPCAddr)

	// 2. Kết nối CSDL & Chạy Migration
	config.ConnectDB()

	// 3. Kết nối hạ tầng RabbitMQ & Redis (hoạt động chế độ fallback nếu lỗi)
	eventPublisher := messaging.NewRabbitPublisher(config.AppConfig)
	defer eventPublisher.Close()
	config.InitRedis()

	// 4. Khởi tạo các tầng nghiệp vụ
	// Repositories
	walletRepo := repository.NewWalletRepository(config.DB)
	paymentRepo := repository.NewPaymentRepository(config.DB)
	withdrawalRepo := repository.NewWithdrawalRepository(config.DB)
	outboxRepo := repository.NewOutboxRepository(config.DB)

	// Usecases
	walletUsecase := appwallet.NewUsecase(walletRepo)

	vnpTmnCode := os.Getenv("VNP_TMN_CODE")
	vnpHashSecret := os.Getenv("VNP_HASH_SECRET")
	vnpPaymentURL := os.Getenv("VNP_PAYMENT_URL")
	vnpReturnURL := os.Getenv("VNP_RETURN_URL")
	vnpayClient := vnpay.NewVNPayClient(vnpTmnCode, vnpHashSecret, vnpPaymentURL, vnpReturnURL)

	paymentUoW := repository.NewUnitOfWork(config.DB, walletUsecase)
	paymentUsecase := apppayment.NewUsecaseWithOptions(paymentRepo, paymentUoW, vnpayClient, bookingSvcClient, apppayment.Options{
		OrderTTL:       config.AppConfig.PaymentOrderTTL,
		MinimumWindow:  config.AppConfig.PaymentMinUsableWindow,
		ManagedExperts: profileClient,
		Profiles:       profileClient,
		Appointments:   bookingSvcClient,
	})
	withdrawalUsecase := appwithdrawal.NewUsecase(withdrawalRepo, walletUsecase)

	// Handlers
	wHandler := walletHandler.NewHandler(walletUsecase).WithManagedReader(appwallet.NewManagedReader(walletRepo, profileClient))
	pHandler := paymentHandler.NewHandler(paymentUsecase)
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
	walletWorker := appwallet.NewWorker(config.DB, walletRepo, holdDuration)
	go walletWorker.Start(ctx)

	// Worker 2: Xử lý timeout của các yêu cầu rút PROCESSING quá 5 phút
	withdrawalWorker := appwithdrawal.NewWorker(config.DB, withdrawalUsecase, 5*time.Minute)
	go withdrawalWorker.Start(ctx)

	// Worker 3: Quét Outbox events để dispatch (gRPC sang Booking + event payment.* lên RabbitMQ)
	outboxPublisher := messaging.NewPublisherWithOptions(outboxRepo, bookingSvcClient, messaging.PublisherOptions{
		PollInterval: config.AppConfig.OutboxPollInterval,
		MaxAttempts:  config.AppConfig.OutboxMaxAttempts,
		BaseBackoff:  config.AppConfig.OutboxBaseBackoff,
		MaxBackoff:   config.AppConfig.OutboxMaxBackoff,
		BatchSize:    config.AppConfig.OutboxBatchSize,
		Events:       eventPublisher,
	})
	go outboxPublisher.Start(ctx)

	// Worker 4: Đối soát ví (Ledger Audit) & Đối soát giao dịch VNPay
	reconciliationWorker := appwallet.NewReconciliationWorker(config.DB)
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
