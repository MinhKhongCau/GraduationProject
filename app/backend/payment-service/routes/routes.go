package routes

import (
	paymentHandler "payment-service/internal/payment/handler"
	walletHandler "payment-service/internal/wallet/handler"
	withdrawalHandler "payment-service/internal/withdrawal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	wHandler *walletHandler.Handler,
	pHandler *paymentHandler.Handler,
	wdHandler *withdrawalHandler.Handler,
) {
	api := r.Group("/api/v1/payments")
	{
		// Ví điện tử
		wallets := api.Group("/wallets")
		{
			wallets.GET("/me", wHandler.GetWallet)
			wallets.GET("/history", wHandler.GetHistory)
			wallets.POST("/top-up", wHandler.TopUpWallet)
		}

		// Đơn hàng thanh toán & VNPay Webhook
		api.POST("/orders", pHandler.CreateOrder)
		api.GET("/vnpay-ipn", pHandler.HandleVNPayIPN)

		// Tài khoản ngân hàng
		api.POST("/bank-accounts", wdHandler.LinkBankAccount)
		api.GET("/bank-accounts", wdHandler.GetBankAccounts)

		// Phiếu rút tiền
		api.POST("/withdrawals", wdHandler.CreateWithdrawal)
		api.POST("/withdrawals/:id/approve", wdHandler.ApproveWithdrawal)
		api.POST("/withdrawals/:id/reject", wdHandler.RejectWithdrawal)
	}
}
