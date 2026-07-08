package slot

import "errors"

var (
	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotUnavailable   = errors.New("slot is not available")
	ErrSlotAlreadyLocked = errors.New("slot is already locked by someone else")
)
