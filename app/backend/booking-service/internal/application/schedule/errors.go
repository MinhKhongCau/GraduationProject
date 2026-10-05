package schedule

import "errors"

var (
	ErrTemplateNotFound     = errors.New("time template not found")
	ErrAvailabilityNotFound = errors.New("availability not found")
	ErrInvalidTimeRange     = errors.New("invalid time range")
	ErrNotFound             = errors.New("schedule record not found")
)
