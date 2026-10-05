package timeoff

import "testing"

func TestValidateTimeOffRange(t *testing.T) {
	for _, tt := range []struct {
		name       string
		start, end int64
		valid      bool
	}{
		{"future", 200, 300, true},
		{"equal", 200, 200, false},
		{"reversed", 300, 200, false},
		{"entirely past", 1, 99, false},
		{"missing start", 0, 300, false},
		{"missing end", 200, 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeOffRange(tt.start, tt.end, 100)
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v, err=%v", tt.valid, err)
			}
		})
	}
}

func TestTimeOffCoverageUsesHalfOpenIntervals(t *testing.T) {
	off := ExpertTimeOff{StartDatetime: 100, EndDatetime: 200}
	for _, tt := range []struct {
		name       string
		start, end int64
		covered    bool
	}{
		{"ends at start", 50, 100, false},
		{"starts at end", 200, 250, false},
		{"partial", 50, 101, true},
		{"time-off inside slot", 50, 250, true},
		{"slot inside time-off", 120, 180, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCoveredByTimeOff(tt.start, tt.end, off); got != tt.covered {
				t.Fatalf("covered=%v, got %v", tt.covered, got)
			}
		})
	}
}

func TestProcessedTimeOffRemainsActiveCoverage(t *testing.T) {
	processedAt := int64(500)
	off := ExpertTimeOff{StartDatetime: 100, EndDatetime: 200, ProcessedAt: &processedAt}
	if !IsCoveredByTimeOff(120, 180, off) {
		t.Fatal("processed_at must not disable time-off coverage")
	}
}
