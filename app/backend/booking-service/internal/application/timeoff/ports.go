package timeoff

import (
	timeoffdomain "booking-service/internal/domain/timeoff"
	"time"
)

type Repository interface {
	CreateAndProcessTimeOff(timeOff *timeoffdomain.ExpertTimeOff, force bool, nowMs int64) ([]string, error)
	ProcessExistingTimeOff(timeOffID string, force bool, nowMs int64) error
	GetUnprocessedTimeOffs() ([]timeoffdomain.ExpertTimeOff, error)
	GetTimeOffsByExpert(expertID string) ([]timeoffdomain.ExpertTimeOff, error)
	DeleteTimeOff(timeOffID string, expertID string) error
	GetTimeOffs(expertID string, fromDate time.Time) ([]timeoffdomain.ExpertTimeOff, error)
}
