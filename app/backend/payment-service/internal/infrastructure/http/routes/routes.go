package routes

import (
	paymentHandler "payment-service/internal/infrastructure/http/handlers/payment"
	walletHandler "payment-service/internal/infrastructure/http/handlers/wallet"
	withdrawalHandler "payment-service/internal/infrastructure/http/handlers/withdrawal"
	"payment-service/internal/infrastructure/http/middleware"

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
		api.GET("/orders/summary", pHandler.SummarizePaymentOrders)
		api.GET("/orders/:id", pHandler.GetPaymentOrder)
		api.GET("/vnpay-ipn", pHandler.HandleVNPayIPN)
		api.GET("/vnpay-return", pHandler.HandleVNPayReturn)

		// Quản lý giao dịch: chuyên gia xem doanh thu của mình
		api.GET("/expert/orders", pHandler.ListExpertOrders)
		api.GET("/expert/orders/summary", pHandler.SummarizeExpertOrders)
		api.GET("/expert/orders/:id", pHandler.GetExpertOrder)

		// Quản lý giao dịch: Admin chỉ thấy/xử lý giao dịch của chuyên gia mình đã duyệt
		api.GET("/admin/orders", pHandler.ListAdminOrders)
		api.GET("/admin/orders/summary", pHandler.SummarizeAdminOrders)
		api.GET("/admin/orders/:id", pHandler.GetAdminOrder)
		api.POST("/admin/orders/:id/review", pHandler.ReviewOrder)
		api.GET("/admin/wallet-transactions", wHandler.ListManagedTransactions)
		api.GET("/compensation-cases", pHandler.ListCompensationCases)
		api.GET("/compensation-cases/:id", pHandler.GetCompensationCase)
		api.POST("/compensation-cases/:id/resolve", pHandler.ResolveCompensationCase)

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

	// Route nội bộ (Kong không public /internal): booking-service yêu cầu chi trả sau buổi tư vấn.
	internal := r.Group("/internal/payments", middleware.Middleware())
	{
		internal.POST("/appointments/:id/settle", pHandler.SettleSession)
	}
}
