package appointment

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	appointmentdomain "booking-service/internal/domain/appointment"
)

var (
	ErrMeetingNotFound        = errors.New("meeting not found")
	ErrSessionForbidden       = errors.New("only the appointment's expert can complete the session")
	ErrSessionNotCompletable  = errors.New("appointment cannot complete a session")
	ErrSessionNotStarted      = errors.New("session has not started yet")
	ErrSessionTooShort        = errors.New("expert presence is shorter than the required minimum")
	ErrSettlementUnavailable  = errors.New("session settlement is not configured")
	ErrSessionSettlementFails = errors.New("session completed but payout failed")
)

const DefaultMeetingBaseURL = "http://localhost:3000/meet"

// SessionSettlementGateway yêu cầu payment-service chi trả thù lao chuyên gia cho lịch hẹn.
// Phải idempotent: booking gọi lại khi lần trước thất bại.
type SessionSettlementGateway interface {
	SettleSession(ctx context.Context, appointmentID string) error
}

// MeetingRepository tra cứu phòng họp theo mã trong link.
type MeetingRepository interface {
	GetMeetingByToken(ctx context.Context, token string) (*MeetingInfo, error)
}

type MeetingOptions struct {
	// BaseURL: link phòng họp = BaseURL + "/" + token (trang phòng họp của frontend).
	BaseURL         string
	MinimumPresence time.Duration
	Settlement      SessionSettlementGateway
}

// MeetingInfo là thông tin chatroom-service cần để cho phép vào phòng họp.
type MeetingInfo struct {
	AppointmentID      string `json:"appointment_id"`
	PatientID          string `json:"patient_id"`
	ExpertID           string `json:"expert_id"`
	Status             string `json:"status"`
	StartTime          int64  `json:"start_time"`
	EndTime            int64  `json:"end_time"`
	SessionCompletedAt *int64 `json:"session_completed_at"`
	// MinimumPresenceSeconds: chatroom-service dùng để biết khi nào chuyên gia được kết thúc buổi.
	MinimumPresenceSeconds int64 `json:"minimum_presence_seconds"`
}

type CompleteSessionCommand struct {
	AppointmentID string
	ExpertID      string
	Presence      time.Duration
}

type SessionCompletion struct {
	AppointmentID      string `json:"appointment_id"`
	SessionCompletedAt int64  `json:"session_completed_at"`
	AlreadyCompleted   bool   `json:"already_completed"`
	Settled            bool   `json:"settled"`
}

// MeetingUsecase: phòng họp giữa bệnh nhân và chuyên gia (chatroom-service gọi qua /internal).
type MeetingUsecase interface {
	GetMeetingByToken(ctx context.Context, token string) (*MeetingInfo, error)
	CompleteSession(ctx context.Context, command CompleteSessionCommand) (*SessionCompletion, error)
}

// NewUsecaseWithMeetings giống NewUsecaseWithProfiles và bật phòng họp + chi trả sau buổi tư vấn.
func NewUsecaseWithMeetings(repo Repository, profiles ProfileGateway, options MeetingOptions) Usecase {
	usecase := NewUsecaseWithProfiles(repo, profiles).(*appointmentUsecase)
	usecase.meetings = normalizeMeetingOptions(options)
	return usecase
}

func normalizeMeetingOptions(options MeetingOptions) MeetingOptions {
	options.BaseURL = strings.TrimRight(strings.TrimSpace(options.BaseURL), "/")
	if options.BaseURL == "" {
		options.BaseURL = DefaultMeetingBaseURL
	}
	if options.MinimumPresence <= 0 {
		options.MinimumPresence = appointmentdomain.DefaultMinimumExpertPresence
	}
	return options
}

func newMeetingToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate meeting token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// addMeetingLink gắn mã + link phòng họp vào bản cập nhật xác nhận thanh toán.
func (u *appointmentUsecase) addMeetingLink(updates map[string]interface{}) error {
	token, err := newMeetingToken()
	if err != nil {
		return err
	}
	updates["meeting_token"] = token
	updates["meeting_link"] = u.meetings.BaseURL + "/" + token
	return nil
}

func (u *appointmentUsecase) GetMeetingByToken(ctx context.Context, token string) (*MeetingInfo, error) {
	token = strings.TrimSpace(token)
	repo, ok := u.repo.(MeetingRepository)
	if token == "" || !ok {
		return nil, ErrMeetingNotFound
	}
	info, err := repo.GetMeetingByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	info.MinimumPresenceSeconds = int64(u.meetings.MinimumPresence / time.Second)
	return info, nil
}

// CompleteSession kết thúc buổi tư vấn (chuyển lịch hẹn → COMPLETED) rồi yêu cầu payment-service
// chuyển thù lao vào ví chuyên gia. Gọi lại an toàn: nếu buổi đã kết thúc thì chỉ chi trả lại.
func (u *appointmentUsecase) CompleteSession(ctx context.Context, command CompleteSessionCommand) (*SessionCompletion, error) {
	if command.AppointmentID == "" {
		return nil, ErrNotFound
	}
	if u.uow == nil {
		return nil, ErrReadRepositoryUnavailable
	}
	if u.meetings.Settlement == nil {
		return nil, ErrSettlementUnavailable
	}

	result := &SessionCompletion{AppointmentID: command.AppointmentID}
	err := u.uow.WithinTx(ctx, func(tx Tx) error {
		appt, err := tx.LoadAppointmentForUpdate(ctx, command.AppointmentID)
		if err != nil {
			return err
		}
		slot, err := tx.LoadSlotForUpdate(ctx, appt.SlotID)
		if err != nil {
			return err
		}
		now := time.Now().UnixMilli()
		transition, err := appointmentdomain.PlanSessionCompletion(*appt, *slot, command.ExpertID, command.Presence, u.meetings.MinimumPresence, now)
		if err != nil {
			return mapSessionDomainError(err)
		}
		if transition.AlreadyCompleted {
			result.AlreadyCompleted = true
			result.SessionCompletedAt = *appt.SessionCompletedAt
			return nil
		}
		result.SessionCompletedAt = now
		return tx.UpdateAppointment(ctx, appt, transition.AppointmentUpdates)
	})
	if err != nil {
		return nil, err
	}

	if err := u.meetings.Settlement.SettleSession(ctx, command.AppointmentID); err != nil {
		return result, fmt.Errorf("%w: %v", ErrSessionSettlementFails, err)
	}
	result.Settled = true
	return result, nil
}

func mapSessionDomainError(err error) error {
	switch {
	case errors.Is(err, appointmentdomain.ErrSessionForbidden):
		return fmt.Errorf("%w: %v", ErrSessionForbidden, err)
	case errors.Is(err, appointmentdomain.ErrSessionNotCompletable):
		return fmt.Errorf("%w: %v", ErrSessionNotCompletable, err)
	case errors.Is(err, appointmentdomain.ErrSessionNotStarted):
		return fmt.Errorf("%w: %v", ErrSessionNotStarted, err)
	case errors.Is(err, appointmentdomain.ErrSessionTooShort):
		return fmt.Errorf("%w: %v", ErrSessionTooShort, err)
	default:
		return err
	}
}
