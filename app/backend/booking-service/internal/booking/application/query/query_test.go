package query

import (
	"errors"
	"testing"
	"time"
)

func TestParsePage(t *testing.T) {
	tests := []struct {
		name, page, size   string
		wantPage, wantSize int
		wantErr            bool
	}{{name: "defaults", wantPage: 0, wantSize: 20}, {name: "page zero", page: "0", size: "1", wantPage: 0, wantSize: 1}, {name: "negative page", page: "-1", wantErr: true}, {name: "zero size", size: "0", wantErr: true}, {name: "oversized", size: "101", wantErr: true}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePage(tt.page, tt.size)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidQuery) {
					t.Fatalf("expected invalid query, got %v", err)
				}
				return
			}
			if err != nil || got.Page != tt.wantPage || got.Size != tt.wantSize {
				t.Fatalf("got %+v err=%v", got, err)
			}
		})
	}
}

func TestParseDateRange(t *testing.T) {
	now := time.Date(2026, 7, 19, 18, 0, 0, 0, Location)
	tests := []struct {
		name, from, to string
		days           int
		wantErr        bool
	}{{name: "default", days: 30}, {name: "from only", from: "2026-07-01", days: 30}, {name: "to only", to: "2026-07-30", days: 30}, {name: "explicit", from: "2026-07-01", to: "2026-07-30", days: 30}, {name: "invalid", from: "07-01-2026", wantErr: true}, {name: "reversed", from: "2026-07-02", to: "2026-07-01", wantErr: true}, {name: "too long", from: "2025-01-01", to: "2026-07-01", wantErr: true}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateRange(tt.from, tt.to, now)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidQuery) {
					t.Fatalf("expected invalid query, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.ToMs-got.FromMs != int64(tt.days)*24*int64(time.Hour/time.Millisecond) {
				t.Fatalf("unexpected range: %+v", got)
			}
		})
	}
}

func TestNewPageMetadata(t *testing.T) {
	page := NewPage([]int{3, 4}, PageRequest{Page: 1, Size: 2}, 5)
	if page.TotalPages != 3 || !page.HasNext || !page.HasPrevious || page.TotalItems != 5 {
		t.Fatalf("unexpected metadata: %+v", page)
	}
	empty := NewPage[int](nil, PageRequest{Page: 3, Size: 20}, 0)
	if empty.Items == nil || empty.TotalPages != 0 || empty.HasNext {
		t.Fatalf("unexpected empty page: %+v", empty)
	}
}
