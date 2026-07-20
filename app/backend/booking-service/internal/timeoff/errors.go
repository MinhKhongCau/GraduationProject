package timeoff

import "errors"

var (
	ErrTimeOffNotFound = errors.New("time-off not found")
	ErrConflict        = errors.New("time-off conflicts with protected booking")
	ErrInvalidDate     = errors.New("invalid time-off dates")
	ErrDuplicate       = errors.New("time-off overlaps an active time-off")
)
