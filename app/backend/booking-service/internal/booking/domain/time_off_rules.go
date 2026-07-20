package domain

import "errors"

var ErrInvalidTimeOffRange = errors.New("invalid time-off range")

func ValidateTimeOffRange(startMs, endMs, nowMs int64) error {
	if startMs <= 0 || endMs <= 0 || startMs >= endMs || endMs <= nowMs {
		return ErrInvalidTimeOffRange
	}
	return nil
}

func IsCoveredByTimeOff(startMs, endMs int64, timeOff ExpertTimeOff) bool {
	return IntervalsOverlap(startMs, endMs, timeOff.StartDatetime, timeOff.EndDatetime)
}
