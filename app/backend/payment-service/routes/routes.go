package routes

import (
	"payment-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/payments")
	{
		wallets := api.Group("/wallets")
		{
			wallets.POST("/init", handlers.InitWallet)
			wallets.GET("/:owner_id", handlers.GetWallet)
			wallets.POST("/:owner_id/top-up", handlers.TopUpWallet)
			wallets.POST("/pay", handlers.ProcessPayment)
			wallets.POST("/:owner_id/withdraw", handlers.RequestWithdrawal)
			wallets.POST("/withdrawals/:request_id/process", handlers.ProcessWithdrawal)
		}
	}
}
