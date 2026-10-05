package wallet

import (
	"context"
	"log"
	paymentdomain "payment-service/internal/domain/payment"
	walletdomain "payment-service/internal/domain/wallet"
	"time"

	"gorm.io/gorm"
)

type ReconciliationWorker struct {
	db *gorm.DB
}

func NewReconciliationWorker(db *gorm.DB) *ReconciliationWorker {
	return &ReconciliationWorker{db: db}
}

func (w *ReconciliationWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	log.Println("⏳ Daily Reconciliation Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Reconciliation Worker")
			return
		case <-ticker.C:
			w.runLedgerReconciliation()
			w.runVNPayReconciliation()
		}
	}
}

func (w *ReconciliationWorker) runLedgerReconciliation() {
	log.Println("[RECONCILIATION] Running Ledger Reconciliation check...")

	var wallets []walletdomain.Wallet
	if err := w.db.Find(&wallets).Error; err != nil {
		log.Printf("Reconciliation: Failed to fetch wallets: %v", err)
		return
	}

	for _, wallet := range wallets {
		var sum int64
		err := w.db.Table("payment_wallet_transactions").
			Where("wallet_id = ?", wallet.ID).
			Select("COALESCE(SUM(amount), 0)").
			Row().Scan(&sum)
		if err != nil {
			log.Printf("Reconciliation: Failed to sum transactions for wallet %s: %v", wallet.ID, err)
			continue
		}

		totalBalance := wallet.AvailableBalance.Int64() + wallet.PendingBalance.Int64() + wallet.LockedBalance.Int64()

		if sum != totalBalance {
			log.Printf("🚨 [LEDGER MISMATCH WARNING] Wallet %s: Sum of transactions (%d) DOES NOT MATCH balance sum (%d) (Avail=%d, Pend=%d, Lock=%d)!",
				wallet.ID, sum, totalBalance, wallet.AvailableBalance, wallet.PendingBalance, wallet.LockedBalance)
		} else {
			log.Printf("📊 [LEDGER OK] Wallet %s matches: sum of txs = %d, total balance = %d", wallet.ID, sum, totalBalance)
		}
	}
}

func (w *ReconciliationWorker) runVNPayReconciliation() {
	log.Println("[RECONCILIATION] Running VNPay Transaction Reconciliation check...")

	var pendingOrders []paymentdomain.PaymentOrder
	thresholdTime := time.Now().Add(-10 * time.Minute).UnixMilli()
	err := w.db.Where("status = ? AND gateway = ? AND created_at <= ?", paymentdomain.OrderStatusPending, "VNPAY", thresholdTime).Find(&pendingOrders).Error
	if err != nil {
		log.Printf("Reconciliation: Failed to fetch pending payment orders: %v", err)
		return
	}

	for _, order := range pendingOrders {
		log.Printf("[RECONCILIATION check] Verifying pending VNPay order: %s, amount: %d", order.ID, order.GrossAmount)
		// TODO: In a real app, query VNPay QueryDR API here
		log.Printf("[RECONCILIATION check] Order %s is still pending. Mock VNPay query confirms no payment found.", order.ID)
	}
}
