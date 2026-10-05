package schedule

import (
	"booking-service/internal/domain/shared"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	MinSlotDurationMinutes = 10
	MaxSlotDurationMinutes = 180
)

var (
	ErrInvalidSchedule     = errors.New("invalid schedule configuration")
	ErrScheduleOverlap     = errors.New("schedule configuration overlaps")
	strictClockTimePattern = regexp.MustCompile(`^(?:[01][0-9]|2[0-3]):[0-5][0-9]$`)
)

type ClockTime struct {
	Hour   int
	Minute int
}

func ParseClockTime(value string) (ClockTime, error) {
	if !strictClockTimePattern.MatchString(value) {
		return ClockTime{}, fmt.Errorf("%w: time %q must use HH:mm", ErrInvalidSchedule, value)
	}
	hour, _ := strconv.Atoi(value[:2])
	minute, _ := strconv.Atoi(value[3:])
	return ClockTime{Hour: hour, Minute: minute}, nil
}

func (t ClockTime) MinutesSinceMidnight() int { return t.Hour*60 + t.Minute }

func ValidateTimeTemplate(template TimeTemplate) error {
	start, err := ParseClockTime(template.StartTime)
	if err != nil {
		return err
	}
	end, err := ParseClockTime(template.EndTime)
	if err != nil {
		return err
	}
	if start.MinutesSinceMidnight() >= end.MinutesSinceMidnight() {
		return fmt.Errorf("%w: template start must be before end", ErrInvalidSchedule)
	}
	if template.SlotDurationMinutes < MinSlotDurationMinutes || template.SlotDurationMinutes > MaxSlotDurationMinutes {
		return fmt.Errorf("%w: slot duration must be between %d and %d minutes", ErrInvalidSchedule, MinSlotDurationMinutes, MaxSlotDurationMinutes)
	}
	return nil
}

func ValidateAvailability(availability Availability, template TimeTemplate) (shared.MoneyVND, error) {
	if strings.TrimSpace(availability.ExpertID) == "" {
		return 0, fmt.Errorf("%w: expert is required", ErrInvalidSchedule)
	}
	if availability.DayOfWeek < 1 || availability.DayOfWeek > 7 {
		return 0, fmt.Errorf("%w: day_of_week must be between 1 and 7", ErrInvalidSchedule)
	}
	if availability.EffectiveFrom <= 0 {
		return 0, fmt.Errorf("%w: effective_from is required", ErrInvalidSchedule)
	}
	if availability.EffectiveUntil != nil && *availability.EffectiveUntil < availability.EffectiveFrom {
		return 0, fmt.Errorf("%w: effective_until must not be before effective_from", ErrInvalidSchedule)
	}
	if err := ValidateTimeTemplate(template); err != nil {
		return 0, err
	}
	if availability.IsEnabled && !template.IsActive {
		return 0, fmt.Errorf("%w: enabled availability requires an active template", ErrInvalidSchedule)
	}
	if availability.Price == nil {
		return 0, fmt.Errorf("%w: availability price is required", shared.ErrInvalidMoneyVND)
	}
	return shared.NewMoneyVNDFromPrice(*availability.Price)
}

func AvailabilityAppliesOnDate(availability Availability, target time.Time, location *time.Location) bool {
	if !availability.IsEnabled {
		return false
	}
	targetDay := calendarDay(target, location)
	fromDay := calendarDay(time.UnixMilli(availability.EffectiveFrom), location)
	if targetDay.Before(fromDay) {
		return false
	}
	if availability.EffectiveUntil != nil {
		untilDay := calendarDay(time.UnixMilli(*availability.EffectiveUntil), location)
		if targetDay.After(untilDay) {
			return false
		}
	}
	return true
}

func AvailabilitiesOverlap(left Availability, leftTemplate TimeTemplate, right Availability, rightTemplate TimeTemplate, location *time.Location) (bool, error) {
	if !left.IsEnabled || !right.IsEnabled || left.DayOfWeek != right.DayOfWeek {
		return false, nil
	}
	leftStart, err := ParseClockTime(leftTemplate.StartTime)
	if err != nil {
		return false, err
	}
	leftEnd, err := ParseClockTime(leftTemplate.EndTime)
	if err != nil {
		return false, err
	}
	rightStart, err := ParseClockTime(rightTemplate.StartTime)
	if err != nil {
		return false, err
	}
	rightEnd, err := ParseClockTime(rightTemplate.EndTime)
	if err != nil {
		return false, err
	}
	if leftStart.MinutesSinceMidnight() >= rightEnd.MinutesSinceMidnight() || leftEnd.MinutesSinceMidnight() <= rightStart.MinutesSinceMidnight() {
		return false, nil
	}
	return effectiveRangesOverlap(left, right, location), nil
}

func effectiveRangesOverlap(left, right Availability, location *time.Location) bool {
	leftStart := calendarDay(time.UnixMilli(left.EffectiveFrom), location)
	rightStart := calendarDay(time.UnixMilli(right.EffectiveFrom), location)
	leftEnd := farFutureDay(location)
	rightEnd := farFutureDay(location)
	if left.EffectiveUntil != nil {
		leftEnd = calendarDay(time.UnixMilli(*left.EffectiveUntil), location)
	}
	if right.EffectiveUntil != nil {
		rightEnd = calendarDay(time.UnixMilli(*right.EffectiveUntil), location)
	}
	return !leftStart.After(rightEnd) && !rightStart.After(leftEnd)
}

func calendarDay(value time.Time, location *time.Location) time.Time {
	local := value.In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
}

func farFutureDay(location *time.Location) time.Time {
	return time.Date(9999, 12, 31, 0, 0, 0, 0, location)
}
