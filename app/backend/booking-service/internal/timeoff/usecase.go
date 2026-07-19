package timeoff

import (
	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/booking/domain"
	"booking-service/internal/slot"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type Usecase interface {
	CreateTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*domain.ExpertTimeOff, []string, error)
	ConfirmTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*domain.ExpertTimeOff, error)
	GetTimeOffs(expertID string) ([]domain.ExpertTimeOff, error)
	DeleteTimeOff(expertID, timeOffID string) error
	ProcessTimeOffs()
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

func (u *timeoffUsecase) CreateTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*domain.ExpertTimeOff, []string, error) {
	return u.create(expertID, startDatetime, endDatetime, reason, false)
}

func (u *timeoffUsecase) ConfirmTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*domain.ExpertTimeOff, error) {
	timeOff, _, err := u.create(expertID, startDatetime, endDatetime, reason, true)
	return timeOff, err
}

func (u *timeoffUsecase) create(expertID string, startDatetime, endDatetime int64, reason string, force bool) (*domain.ExpertTimeOff, []string, error) {
	nowMs := u.clock().UnixMilli()
	if err := domain.ValidateTimeOffRange(startDatetime, endDatetime, nowMs); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidDate, err)
	}
	timeOff := &domain.ExpertTimeOff{TimeOffID: uuid.New().String(), ExpertID: expertID, StartDatetime: startDatetime, EndDatetime: endDatetime, Reason: reason}
	conflicts, err := u.repo.CreateAndProcessTimeOff(timeOff, force, nowMs)
	if err != nil {
		return nil, conflicts, err
	}
	return timeOff, nil, nil
}

func (u *timeoffUsecase) GetTimeOffs(expertID string) ([]domain.ExpertTimeOff, error) {
	return u.repo.GetTimeOffsByExpert(expertID)
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
