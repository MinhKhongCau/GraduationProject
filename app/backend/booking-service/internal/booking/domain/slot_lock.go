package domain

const SlotLockDurationMs int64 = 15 * 60 * 1000

type SlotLock struct {
	LockedBy        string
	LockedExpiresAt int64
}

func NewSlotLock(patientID string, nowMs int64) SlotLock {
	return SlotLock{
		LockedBy:        patientID,
		LockedExpiresAt: nowMs + SlotLockDurationMs,
	}
}

func (lock SlotLock) Updates() map[string]interface{} {
	return map[string]interface{}{
		"status":            SlotStatusLocked,
		"locked_expires_at": lock.LockedExpiresAt,
		"locked_by":         lock.LockedBy,
	}
}
