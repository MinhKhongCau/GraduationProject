package appointment

import (
	"booking-service/internal/domain"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PaymentResultStatus string

const (
	PaymentResultSuccess PaymentResultStatus = "SUCCESS"
	PaymentResultFailed  PaymentResultStatus = "FAILED"
)

type HandlePaymentResultCommand struct {
	AppointmentID string
	Status        PaymentResultStatus
}

type GetPaymentEligibilityCommand struct {
	AppointmentID string
	PayerID       string
}

type PaymentEligibility struct {
	AppointmentID string
	ExpertID      string
	AmountVND     int64
	ExpiresAt     int64
}

type Usecase interface {
	CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error)
	GetAppointmentByID(appointmentID string) (*domain.Appointment, error)
	GetPaymentEligibility(command GetPaymentEligibilityCommand) (PaymentEligibility, error)
	CancelAppointment(appointmentID, userID, userRole, reason string) error
	ConfirmPayment(appointmentID string) error
	HandlePaymentFailure(appointmentID string) error
	HandlePaymentResult(command HandlePaymentResultCommand) error
	GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error)
	GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error)
}

type appointmentUsecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &appointmentUsecase{repo: repo}
}

func ParsePaymentResultStatus(status string) (PaymentResultStatus, error) {
	switch PaymentResultStatus(strings.TrimSpace(status)) {
	case PaymentResultSuccess:
		return PaymentResultSuccess, nil
	case PaymentResultFailed:
		return PaymentResultFailed, nil
	default:
		return "", ErrInvalidPaymentResultStatus
	}
}

func (u *appointmentUsecase) CreateAppointment(patientID, expertID, slotID string) (*domain.Appointment, error) {
	appointment := &domain.Appointment{
		AppointmentID: uuid.New().String(),
		SlotID:        slotID,
		PatientID:     patientID,
		ExpertID:      expertID,
		Status:        domain.AppointmentStatusPendingPayment,
	}

	if err := u.repo.CreateAppointment(appointment); err != nil {
		return nil, err
	}
	return appointment, nil
}

func (u *appointmentUsecase) GetAppointmentByID(appointmentID string) (*domain.Appointment, error) {
	return u.repo.GetAppointmentByID(appointmentID)
}

func (u *appointmentUsecase) GetPaymentEligibility(command GetPaymentEligibilityCommand) (PaymentEligibility, error) {
	command.AppointmentID = strings.TrimSpace(command.AppointmentID)
	command.PayerID = strings.TrimSpace(command.PayerID)
	if command.AppointmentID == "" {
		return PaymentEligibility{}, ErrNotFound
	}
	if command.PayerID == "" {
		return PaymentEligibility{}, ErrPaymentEligibilityForbidden
	}

	snapshot, err := u.repo.GetPaymentEligibilitySnapshot(command)
	if err != nil {
		return PaymentEligibility{}, err
	}

	return buildPaymentEligibility(command, snapshot.Appointment, snapshot.Slot, time.Now().UnixMilli())
}

func (u *appointmentUsecase) CancelAppointment(appointmentID, userID, userRole, reason string) error {
	if userRole == "PATIENT" {
		// TODO: Mốc thời gian huỷ tối thiểu (policy)
		// Check nếu cuộc hẹn bắt đầu trong vòng 24h thì chặn không cho huỷ
		return u.repo.CancelAppointmentByPatient(appointmentID, userID, reason)
	} else if userRole == "EXPERT" {
		return u.repo.CancelAppointmentByExpert(appointmentID, reason)
	}
	return ErrUnauthorized
}

func (u *appointmentUsecase) ConfirmPayment(appointmentID string) error {
	return u.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultSuccess,
	})
}

func (u *appointmentUsecase) HandlePaymentFailure(appointmentID string) error {
	return u.HandlePaymentResult(HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        PaymentResultFailed,
	})
}

func (u *appointmentUsecase) HandlePaymentResult(command HandlePaymentResultCommand) error {
	if command.AppointmentID == "" {
		return ErrNotFound
	}
	if command.Status != PaymentResultSuccess && command.Status != PaymentResultFailed {
		return ErrInvalidPaymentResultStatus
	}
	return u.repo.HandlePaymentResult(command)
}

func (u *appointmentUsecase) GetAppointmentsByPatient(patientID string) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByPatient(patientID)
}

func (u *appointmentUsecase) GetAppointmentsByExpert(expertID string, fromDate, toDate int64, status *domain.AppointmentStatus) ([]domain.Appointment, error) {
	return u.repo.GetAppointmentsByExpert(expertID, fromDate, toDate, status)
}

func buildPaymentEligibility(command GetPaymentEligibilityCommand, appt domain.Appointment, slot domain.ExpertSlot, nowMs int64) (PaymentEligibility, error) {
	if !sameID(appt.PatientID, command.PayerID) {
		return PaymentEligibility{}, ErrPaymentEligibilityForbidden
	}
	if appt.Status != domain.AppointmentStatusPendingPayment {
		return PaymentEligibility{}, fmt.Errorf("%w: appointment status is %d", ErrPaymentEligibilityConflict, appt.Status)
	}
	if slot.SlotID == "" {
		return PaymentEligibility{}, fmt.Errorf("%w: slot not found", ErrPaymentEligibilityConflict)
	}
	if !sameID(slot.ExpertID, appt.ExpertID) {
		return PaymentEligibility{}, fmt.Errorf("%w: slot expert does not match appointment expert", ErrPaymentEligibilityConflict)
	}
	if slot.Status != domain.SlotStatusLocked {
		return PaymentEligibility{}, fmt.Errorf("%w: slot status is %d", ErrPaymentEligibilityConflict, slot.Status)
	}
	if slot.LockedBy == nil || !sameID(*slot.LockedBy, command.PayerID) {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock owner mismatch", ErrPaymentEligibilityConflict)
	}
	if slot.LockedExpiresAt == nil {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock expiry is missing", ErrPaymentEligibilityConflict)
	}
	if *slot.LockedExpiresAt <= nowMs {
		return PaymentEligibility{}, fmt.Errorf("%w: slot lock expired", ErrPaymentEligibilityConflict)
	}

	amountVND, err := strictPriceToVND(slot.Price)
	if err != nil {
		return PaymentEligibility{}, err
	}

	return PaymentEligibility{
		AppointmentID: appt.AppointmentID,
		ExpertID:      appt.ExpertID,
		AmountVND:     amountVND,
		ExpiresAt:     *slot.LockedExpiresAt,
	}, nil
}

func strictPriceToVND(price float64) (int64, error) {
	if math.IsNaN(price) {
		return 0, fmt.Errorf("%w: price is NaN", ErrInvalidBookingPrice)
	}
	if math.IsInf(price, 0) {
		return 0, fmt.Errorf("%w: price is infinite", ErrInvalidBookingPrice)
	}
	if price <= 0 {
		return 0, fmt.Errorf("%w: price must be greater than zero", ErrInvalidBookingPrice)
	}
	if math.Trunc(price) != price {
		return 0, fmt.Errorf("%w: fractional VND is not supported", ErrInvalidBookingPrice)
	}
	const maxVNPaySafeAmountVND = float64(92233720368547758)
	if price >= maxVNPaySafeAmountVND {
		return 0, fmt.Errorf("%w: price is too large", ErrInvalidBookingPrice)
	}
	return int64(price), nil
}

func sameID(left, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}
