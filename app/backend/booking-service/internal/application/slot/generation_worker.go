package slot

import (
	"context"
	"log"
	"time"
)

const RollingGenerationInterval = 24 * time.Hour

func RunStartupGeneration(ctx context.Context, app GenerationApplication, days int) GenerationRunSummary {
	log.Printf("[slot-generation] startup run started days=%d", days)
	summary := app.GenerateAll(ctx, days)
	logGenerationSummary("startup", summary)
	return summary
}

func StartRollingGenerationWorker(ctx context.Context, app GenerationApplication, days int) {
	StartRollingGenerationWorkerWithInterval(ctx, app, days, RollingGenerationInterval)
}

func StartRollingGenerationWorkerWithInterval(ctx context.Context, app GenerationApplication, days int, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("[slot-generation] worker stopped")
				return
			case <-ticker.C:
				log.Printf("[slot-generation] daily run started days=%d", days)
				logGenerationSummary("daily", app.GenerateAll(ctx, days))
			}
		}
	}()
}

func logGenerationSummary(kind string, summary GenerationRunSummary) {
	for _, failure := range summary.Errors {
		log.Printf("[slot-generation] %s expert=%s failed: %v", kind, failure.ExpertID, failure.Err)
	}
	log.Printf("[slot-generation] %s run completed processed=%d succeeded=%d failed=%d inserted=%d", kind, summary.Processed, summary.Succeeded, summary.Failed, summary.Inserted)
}
