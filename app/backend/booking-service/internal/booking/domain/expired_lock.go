package domain

const ExpiredPaymentCancellationReason = "QuÃ¡ háº¡n thanh toÃ¡n 15 phÃºt"

type ExpiredLockCleanupPlan struct {
	AppointmentUpdates map[string]interface{}
	SlotUpdates        map[string]interface{}
}

func IsSlotLockExpired(slot ExpertSlot, nowMs int64) bool {
	return slot.Status == SlotStatusLocked && slot.LockedExpiresAt != nil && *slot.LockedExpiresAt < nowMs
}

func PlanExpiredLockCleanup(nowMs int64) ExpiredLockCleanupPlan {
	return ExpiredLockCleanupPlan{
		AppointmentUpdates: map[string]interface{}{
			"status":              AppointmentStatusCancelled,
			"cancellation_reason": ExpiredPaymentCancellationReason,
			"cancelled_by":        cancellationActorPtr(CancellationActorSystem),
			"updated_at":          nowMs,
		},
		SlotUpdates: ReleasedSlotUpdates(),
	}
}

func cancellationActorPtr(actor CancellationActor) *string {
	value := string(actor)
	return &value
}
