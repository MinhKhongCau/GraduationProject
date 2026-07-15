package withdrawal

import (
	"context"
	"log"
	"payment-service/internal/domain/entity"
	"time"

	"gorm.io/gorm"
)

type Worker struct {
	db      *gorm.DB
	usecase Usecase
	timeout time.Duration
}

func NewWorker(db *gorm.DB, usecase Usecase, timeout time.Duration) *Worker {
	return &Worker{
		db:      db,
		usecase: usecase,
		timeout: timeout,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("⏳ Withdrawal Processing Timeout Worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Withdrawal Processing Timeout Worker")
			return
		case <-ticker.C:
			w.checkProcessingTimeouts()
		}
	}
}

func (w *Worker) checkProcessingTimeouts() {
	now := time.Now().UnixMilli()
	timeoutThreshold := now - w.timeout.Milliseconds()

	var stuckRequests []entity.WithdrawalRequest
	err := w.db.Where("status = ? AND requested_at <= ?", entity.WithdrawalStatusProcessing, timeoutThreshold).Find(&stuckRequests).Error
	if err != nil {
		log.Printf("Worker error finding stuck processing requests: %v", err)
		return
	}

	ctx := context.Background()
	for _, req := range stuckRequests {
		log.Printf("Worker: Request %s is stuck in PROCESSING. Marking as FAILED due to timeout.", req.ID)
		err := w.usecase.HandlePayoutCallback(ctx, req.ID, "TIMEOUT_AUTO_ROLLBACK", false, "Payout gateway response timed out")
		if err != nil {
			log.Printf("Worker: Failed to time out stuck request %s: %v", req.ID, err)
		} else {
			log.Printf("Worker: Successfully timed out stuck request %s", req.ID)
		}
	}
}
