package routes

import (
	paymentHandler "payment-service/internal/infrastructure/http/handlers/payment"
	walletHandler "payment-service/internal/infrastructure/http/handlers/wallet"
	withdrawalHandler "payment-service/internal/infrastructure/http/handlers/withdrawal"

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
		api.GET("/orders", pHandler.ListPaymentOrders)
		api.GET("/orders/:id", pHandler.GetPaymentOrder)
		api.GET("/vnpay-ipn", pHandler.HandleVNPayIPN)
		api.GET("/compensation-cases", pHandler.ListCompensationCases)
		api.GET("/compensation-cases/:id", pHandler.GetCompensationCase)

		// Tài khoản ngân hàng
		api.POST("/bank-accounts", wdHandler.LinkBankAccount)
		api.GET("/bank-accounts", wdHandler.GetBankAccounts)

		// Phiếu rút tiền
		api.POST("/withdrawals", wdHandler.CreateWithdrawal)
		api.GET("/withdrawals", wdHandler.ListWithdrawals)
		api.GET("/withdrawals/:id", wdHandler.GetWithdrawal)
		api.GET("/admin/withdrawals", wdHandler.ListAdminWithdrawals)
		api.POST("/withdrawals/:id/approve", wdHandler.ApproveWithdrawal)
		api.POST("/withdrawals/:id/reject", wdHandler.RejectWithdrawal)
	}
}
