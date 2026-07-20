package readquery

import (
	"errors"
	"testing"
	"time"
)

func TestPaginationAndRangeValidation(t *testing.T) {
	if got, err := ParsePage("", ""); err != nil || got.Page != 0 || got.Size != 20 {
		t.Fatalf("defaults: %+v %v", got, err)
	}
	for _, input := range [][2]string{{"-1", "20"}, {"0", "0"}, {"0", "101"}} {
		if _, err := ParsePage(input[0], input[1]); !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected error for %v", input)
		}
	}
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, Location)
	got, err := ParseDateRange("", "", now)
	if err != nil || got.ToMs-got.FromMs != 30*24*int64(time.Hour/time.Millisecond) {
		t.Fatalf("range: %+v %v", got, err)
	}
	if _, err := ParseDateRange("2026-07-02", "2026-07-01", now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected reversed error")
	}
}

func TestPageMetadata(t *testing.T) {
	got := NewPage([]int{1}, PageRequest{Page: 1, Size: 1}, 3)
	if got.TotalPages != 3 || !got.HasNext || !got.HasPrevious {
		t.Fatalf("metadata: %+v", got)
	}
}
