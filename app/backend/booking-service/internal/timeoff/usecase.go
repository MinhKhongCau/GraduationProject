package timeoff

import (
	appappointment "booking-service/internal/booking/application/appointment"
	"booking-service/internal/booking/domain"
	"booking-service/internal/slot"
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
	repo            Repository
	slotRepo        slot.Repository
	appointmentRepo appappointment.Repository
}

func NewUsecase(repo Repository, slotRepo slot.Repository, appointmentRepo appappointment.Repository) Usecase {
	return &timeoffUsecase{
		repo:            repo,
		slotRepo:        slotRepo,
		appointmentRepo: appointmentRepo,
	}
}

func (u *timeoffUsecase) CreateTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*domain.ExpertTimeOff, []string, error) {
	// 1. Conflict detection
	overlappingSlots, err := u.repo.GetOverlappingSlots(expertID, startDatetime, endDatetime)
	if err != nil {
		return nil, nil, err
	}

	hasOccupied := false
	var affectedAppointments []string

	for _, slot := range overlappingSlots {
		if slot.Status == domain.SlotStatusOccupied {
			hasOccupied = true
			// Find affected appointment
			appt, _ := u.appointmentRepo.GetAppointmentBySlotID(slot.SlotID)
			if appt != nil {
				affectedAppointments = append(affectedAppointments, appt.AppointmentID)
			}
		}
	}

	// 2. If there are occupied slots -> Return ErrConflict
	if hasOccupied {
		return nil, affectedAppointments, ErrConflict
	}

	// 3. Save TimeOff
	timeOff := domain.ExpertTimeOff{
		TimeOffID:     uuid.New().String(),
		ExpertID:      expertID,
		StartDatetime: startDatetime,
		EndDatetime:   endDatetime,
		Reason:        reason,
		// ProcessedAt is left empty (null) for the Background Worker to process cancellations
	}

	if err := u.repo.CreateTimeOff(&timeOff); err != nil {
		return nil, nil, err
	}

	return &timeOff, nil, nil
}

func (u *timeoffUsecase) ConfirmTimeOff(expertID string, startDatetime, endDatetime int64, reason string) (*domain.ExpertTimeOff, error) {
	timeOff := domain.ExpertTimeOff{
		TimeOffID:     uuid.New().String(),
		ExpertID:      expertID,
		StartDatetime: startDatetime,
		EndDatetime:   endDatetime,
		Reason:        reason,
	}

	if err := u.repo.CreateTimeOff(&timeOff); err != nil {
		return nil, err
	}

	return &timeOff, nil
}

func (u *timeoffUsecase) GetTimeOffs(expertID string) ([]domain.ExpertTimeOff, error) {
	return u.repo.GetTimeOffsByExpert(expertID)
}

func (u *timeoffUsecase) DeleteTimeOff(expertID, timeOffID string) error {
	// TODO: verify ownership or if it exists
	return u.repo.DeleteTimeOff(timeOffID, expertID)
}

func (u *timeoffUsecase) ProcessTimeOffs() {
	timeOffs, err := u.repo.GetUnprocessedTimeOffs()
	if err != nil {
		return
	}

	for _, to := range timeOffs {
		// 1. Xoá tất cả Slot AVAILABLE trong khoảng thời gian nghỉ
		err := u.repo.DeleteAvailableSlots(to.ExpertID, to.StartDatetime, to.EndDatetime)
		if err != nil {
			continue
		}

		// 2. Tìm tất cả các Appointment bị đè lên
		appts, err := u.repo.GetOverlappingAppointments(to.ExpertID, to.StartDatetime, to.EndDatetime)
		if err != nil {
			continue
		}

		// 3. Huỷ các cuộc hẹn bị đè
		reason := "Bác sĩ đăng ký lịch nghỉ đột xuất: " + to.Reason
		for _, appt := range appts {
			_ = u.appointmentRepo.CancelAppointmentByExpert(appt.AppointmentID, reason)
		}

		// 4. Đánh dấu đã xử lý
		_ = u.repo.MarkAsProcessed(to.TimeOffID)
	}
}
