package timeoff

import (
	appappointment "booking-service/internal/application/appointment"
	bookingquery "booking-service/internal/application/query"
	"booking-service/internal/application/slot"
	timeoffdomain "booking-service/internal/domain/timeoff"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type Usecase interface {
	CreateTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*timeoffdomain.ExpertTimeOff, []string, error)
	ConfirmTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*timeoffdomain.ExpertTimeOff, error)
	GetTimeOffs(expertID string) ([]timeoffdomain.ExpertTimeOff, error)
	ListTimeOffs(query TimeOffListQuery) (bookingquery.Page[timeoffdomain.ExpertTimeOff], error)
	DeleteTimeOff(expertID, timeOffID string) error
	ProcessTimeOffs()
}

type TimeOffListQuery struct {
	ExpertID  string
	FromMs    int64
	ToMs      int64
	Processed *bool
	Page      bookingquery.PageRequest
}
type timeOffReader interface {
	ListTimeOffs(query TimeOffListQuery) ([]timeoffdomain.ExpertTimeOff, int64, error)
}

type timeoffUsecase struct {
	repo  Repository
	clock func() time.Time
}

func NewUsecase(repo Repository, _ slot.Repository, _ appappointment.Repository) Usecase {
	return &timeoffUsecase{repo: repo, clock: time.Now}
}

func newUsecaseWithClock(repo Repository, clock func() time.Time) Usecase {
	return &timeoffUsecase{repo: repo, clock: clock}
}

func (u *timeoffUsecase) CreateTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*timeoffdomain.ExpertTimeOff, []string, error) {
	return u.create(expertID, startDatetime, endDatetime, reason, false)
}

func (u *timeoffUsecase) ConfirmTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*timeoffdomain.ExpertTimeOff, error) {
	timeOff, _, err := u.create(expertID, startDatetime, endDatetime, reason, true)
	return timeOff, err
}

func (u *timeoffUsecase) create(expertID string, startDatetime, endDatetime int64, reason string, force bool) (*timeoffdomain.ExpertTimeOff, []string, error) {
	nowMs := u.clock().UnixMilli()
	if err := timeoffdomain.ValidateTimeOffRange(startDatetime, endDatetime, nowMs); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidDate, err)
	}
	timeOff := &timeoffdomain.ExpertTimeOff{TimeOffID: uuid.New().String(), ExpertID: expertID, StartDatetime: startDatetime, EndDatetime: endDatetime, Reason: reason}
	conflicts, err := u.repo.CreateAndProcessTimeOff(timeOff, force, nowMs)
	if err != nil {
		return nil, conflicts, err
	}
	return timeOff, nil, nil
}

func (u *timeoffUsecase) GetTimeOffs(expertID string) ([]timeoffdomain.ExpertTimeOff, error) {
	return u.repo.GetTimeOffsByExpert(expertID)
}

func (u *timeoffUsecase) ListTimeOffs(filter TimeOffListQuery) (bookingquery.Page[timeoffdomain.ExpertTimeOff], error) {
	reader, ok := u.repo.(timeOffReader)
	if !ok {
		return bookingquery.Page[timeoffdomain.ExpertTimeOff]{}, errors.New("time-off reader unavailable")
	}
	items, total, err := reader.ListTimeOffs(filter)
	if err != nil {
		return bookingquery.Page[timeoffdomain.ExpertTimeOff]{}, err
	}
	return bookingquery.NewPage(items, filter.Page, total), nil
}

func (u *timeoffUsecase) DeleteTimeOff(expertID, timeOffID string) error {
	return u.repo.DeleteTimeOff(timeOffID, expertID)
}

func (u *timeoffUsecase) ProcessTimeOffs() {
	timeOffs, err := u.repo.GetUnprocessedTimeOffs()
	if err != nil {
		log.Printf("[time-off] load unprocessed failed: %v", err)
		return
	}
	for _, item := range timeOffs {
		if err := u.repo.ProcessExistingTimeOff(item.TimeOffID, false, u.clock().UnixMilli()); err != nil {
			log.Printf("[time-off] process id=%s failed: %v", item.TimeOffID, err)
		}
	}
}
