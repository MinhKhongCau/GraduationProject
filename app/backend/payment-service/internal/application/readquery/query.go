package readquery

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DefaultSize = 20
const MaximumSize = 100

var ErrInvalid = errors.New("invalid query")
var Location = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

type PageRequest struct {
	Page int
	Size int
}

func ParsePage(pageValue, sizeValue string) (PageRequest, error) {
	result := PageRequest{Size: DefaultSize}
	var err error
	if value := strings.TrimSpace(pageValue); value != "" {
		result.Page, err = strconv.Atoi(value)
		if err != nil || result.Page < 0 {
			return PageRequest{}, fmt.Errorf("%w: page must be at least 0", ErrInvalid)
		}
	}
	if value := strings.TrimSpace(sizeValue); value != "" {
		result.Size, err = strconv.Atoi(value)
		if err != nil || result.Size < 1 || result.Size > MaximumSize {
			return PageRequest{}, fmt.Errorf("%w: size must be between 1 and %d", ErrInvalid, MaximumSize)
		}
	}
	return result, nil
}
func (p PageRequest) Offset() int { return p.Page * p.Size }

type DateRange struct {
	FromMs int64
	ToMs   int64
}

func ParseDateRange(fromValue, toValue string, now time.Time) (DateRange, error) {
	fromText, toText := strings.TrimSpace(fromValue), strings.TrimSpace(toValue)
	today := time.Date(now.In(Location).Year(), now.In(Location).Month(), now.In(Location).Day(), 0, 0, 0, 0, Location)
	var from, to time.Time
	var err error
	switch {
	case fromText == "" && toText == "":
		from, to = today, today.AddDate(0, 0, 29)
	case fromText != "" && toText == "":
		from, err = parseDate(fromText)
		if err != nil {
			return DateRange{}, err
		}
		to = from.AddDate(0, 0, 29)
	case fromText == "" && toText != "":
		to, err = parseDate(toText)
		if err != nil {
			return DateRange{}, err
		}
		from = to.AddDate(0, 0, -29)
	default:
		from, err = parseDate(fromText)
		if err != nil {
			return DateRange{}, err
		}
		to, err = parseDate(toText)
		if err != nil {
			return DateRange{}, err
		}
	}
	if from.After(to) {
		return DateRange{}, fmt.Errorf("%w: from must not be after to", ErrInvalid)
	}
	toExclusive := to.AddDate(0, 0, 1)
	if int(toExclusive.Sub(from).Hours()/24) > 366 {
		return DateRange{}, fmt.Errorf("%w: date range must not exceed 366 days", ErrInvalid)
	}
	return DateRange{FromMs: from.UnixMilli(), ToMs: toExclusive.UnixMilli()}, nil
}
func parseDate(value string) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", value, Location)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: date must use YYYY-MM-DD", ErrInvalid)
	}
	return parsed, nil
}

type Page[T any] struct {
	Items       []T   `json:"items"`
	Page        int   `json:"page"`
	Size        int   `json:"size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

func NewPage[T any](items []T, request PageRequest, total int64) Page[T] {
	if items == nil {
		items = []T{}
	}
	pages := 0
	if total > 0 {
		pages = int((total + int64(request.Size) - 1) / int64(request.Size))
	}
	return Page[T]{Items: items, Page: request.Page, Size: request.Size, TotalItems: total, TotalPages: pages, HasNext: request.Page+1 < pages, HasPrevious: request.Page > 0}
}
