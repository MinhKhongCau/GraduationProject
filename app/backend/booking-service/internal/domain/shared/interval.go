package shared

func IntervalsOverlap(start, end, otherStart, otherEnd int64) bool {
	return start < otherEnd && end > otherStart
}
