package domain

import "time"

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusRetryWait OutboxStatus = "RETRY_WAIT"
	OutboxStatusDelivered OutboxStatus = "DELIVERED"
	OutboxStatusDead      OutboxStatus = "DEAD"
)

func OutboxRetryDelay(base, maximum time.Duration, attemptCount int) time.Duration {
	if base <= 0 || maximum <= 0 {
		return 0
	}
	if base >= maximum {
		return maximum
	}
	if attemptCount <= 1 {
		return base
	}

	delay := base
	for attempt := 1; attempt < attemptCount; attempt++ {
		if delay >= maximum-delay {
			return maximum
		}
		delay *= 2
	}
	return delay
}

func IsOutboxEligible(status OutboxStatus, nextAttemptAt *int64, nowMillis int64) bool {
	switch status {
	case OutboxStatusPending:
		return true
	case OutboxStatusRetryWait:
		return nextAttemptAt != nil && *nextAttemptAt <= nowMillis
	default:
		return false
	}
}
