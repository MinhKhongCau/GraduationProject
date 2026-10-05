package timeoff

import (
	"context"
	"log"
	"time"
)

const WorkerInterval = 2 * time.Minute

func StartWorker(usecase Usecase) {
	StartWorkerWithContext(context.Background(), usecase, WorkerInterval)
}

func StartWorkerWithContext(ctx context.Context, usecase Usecase, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		log.Println("[time-off] worker started")
		for {
			select {
			case <-ctx.Done():
				log.Println("[time-off] worker stopped")
				return
			case <-ticker.C:
				usecase.ProcessTimeOffs()
			}
		}
	}()
}
