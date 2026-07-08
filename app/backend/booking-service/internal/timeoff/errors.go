package timeoff

import "errors"

var (
	ErrTimeOffNotFound = errors.New("time-off not found")
	ErrConflict        = errors.New("time-off conflicts with existing appointments")
	ErrInvalidDate     = errors.New("invalid time-off dates")
)
