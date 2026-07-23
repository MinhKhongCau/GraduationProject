package domain

import (
	"errors"
	"math"
	"testing"
	"time"
)

var testLocation = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

func TestValidateAvailabilityPrice(t *testing.T) {
	valid := float64(200000)
	template := validTemplate()
	tests := []struct {
		name  string
		price *float64
	}{
		{name: "missing", price: nil},
		{name: "zero", price: floatPtr(0)},
		{name: "negative", price: floatPtr(-1)},
		{name: "fractional", price: floatPtr(1000.5)},
		{name: "nan", price: floatPtr(math.NaN())},
		{name: "positive infinity", price: floatPtr(math.Inf(1))},
		{name: "negative infinity", price: floatPtr(math.Inf(-1))},
		{name: "vnpay overflow", price: floatPtr(92233720368547760)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availability := validAvailability(time.Now(), tt.price)
			if _, err := ValidateAvailability(availability, template); !errors.Is(err, ErrInvalidMoneyVND) {
				t.Fatalf("expected invalid money, got %v", err)
			}
		})
	}
	availability := validAvailability(time.Now(), &valid)
	price, err := ValidateAvailability(availability, template)
	if err != nil || price != 200000 {
		t.Fatalf("expected 200000 VND, got %d, %v", price, err)
	}
}

func TestAvailabilityAppliesOnInclusiveCalendarRange(t *testing.T) {
	from := time.Date(2026, 7, 20, 15, 0, 0, 0, testLocation)
	until := time.Date(2026, 7, 22, 1, 0, 0, 0, testLocation).UnixMilli()
	availability := validAvailability(from, floatPtr(200000))
	availability.EffectiveFrom = from.UnixMilli()
	availability.EffectiveUntil = &until
	tests := []struct {
		name string
		date time.Time
		want bool
	}{
		{"before", time.Date(2026, 7, 19, 23, 0, 0, 0, testLocation), false},
		{"equal from date", time.Date(2026, 7, 20, 0, 0, 0, 0, testLocation), true},
		{"equal until date", time.Date(2026, 7, 22, 23, 0, 0, 0, testLocation), true},
		{"after", time.Date(2026, 7, 23, 0, 0, 0, 0, testLocation), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AvailabilityAppliesOnDate(availability, tt.date, testLocation); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
	availability.EffectiveUntil = nil
	if !AvailabilityAppliesOnDate(availability, time.Date(2030, 1, 1, 0, 0, 0, 0, testLocation), testLocation) {
		t.Fatal("open-ended availability should apply")
	}
	availability.IsEnabled = false
	if AvailabilityAppliesOnDate(availability, from, testLocation) {
		t.Fatal("disabled availability must not apply")
	}
}

func TestValidateTimeTemplateStrictRules(t *testing.T) {
	tests := []TimeTemplate{
		{StartTime: "8:03", EndTime: "10:00", SlotDurationMinutes: 30},
		{StartTime: "25:00", EndTime: "26:00", SlotDurationMinutes: 30},
		{StartTime: "08:00x", EndTime: "10:00", SlotDurationMinutes: 30},
		{StartTime: "08:00", EndTime: "08:00", SlotDurationMinutes: 30},
		{StartTime: "10:00", EndTime: "08:00", SlotDurationMinutes: 30},
		{StartTime: "08:00", EndTime: "10:00", SlotDurationMinutes: 0},
		{StartTime: "08:00", EndTime: "10:00", SlotDurationMinutes: -1},
	}
	for _, template := range tests {
		if err := ValidateTimeTemplate(template); !errors.Is(err, ErrInvalidSchedule) {
			t.Fatalf("expected invalid template %#v, got %v", template, err)
		}
	}
	if err := ValidateTimeTemplate(validTemplate()); err != nil {
		t.Fatalf("valid template rejected: %v", err)
	}
}

func TestAvailabilityOverlapIncludesEffectiveDateRange(t *testing.T) {
	left := validAvailability(time.Date(2026, 7, 20, 0, 0, 0, 0, testLocation), floatPtr(200000))
	left.DayOfWeek = 1
	right := left
	right.AvailabilityID = "right"
	leftTemplate := validTemplate()
	rightTemplate := TimeTemplate{StartTime: "08:30", EndTime: "09:30", SlotDurationMinutes: 30, IsActive: true}
	overlap, err := AvailabilitiesOverlap(left, leftTemplate, right, rightTemplate, testLocation)
	if err != nil || !overlap {
		t.Fatalf("expected overlap, got %v, %v", overlap, err)
	}
	leftUntil := time.Date(2026, 7, 31, 0, 0, 0, 0, testLocation).UnixMilli()
	left.EffectiveUntil = &leftUntil
	right.EffectiveFrom = time.Date(2026, 8, 1, 0, 0, 0, 0, testLocation).UnixMilli()
	overlap, err = AvailabilitiesOverlap(left, leftTemplate, right, rightTemplate, testLocation)
	if err != nil || overlap {
		t.Fatalf("non-overlapping effective ranges should be allowed")
	}
	adjacent := TimeTemplate{StartTime: "09:00", EndTime: "10:00", SlotDurationMinutes: 30, IsActive: true}
	right.EffectiveFrom = left.EffectiveFrom
	overlap, _ = AvailabilitiesOverlap(left, leftTemplate, right, adjacent, testLocation)
	if overlap {
		t.Fatal("adjacent template intervals should be allowed")
	}
}

func validTemplate() TimeTemplate {
	return TimeTemplate{TemplateID: "tpl", StartTime: "08:00", EndTime: "09:00", SlotDurationMinutes: 30, IsActive: true}
}
func validAvailability(from time.Time, price *float64) Availability {
	return Availability{AvailabilityID: "avail", ExpertID: "expert", TemplateID: "tpl", DayOfWeek: 1, IsEnabled: true, EffectiveFrom: from.UnixMilli(), Price: price}
}
func floatPtr(value float64) *float64 { return &value }
