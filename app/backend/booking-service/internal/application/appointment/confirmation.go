package appointment

import (
	"context"
	"errors"
	"time"

	appointmentdomain "booking-service/internal/domain/appointment"
	slotdomain "booking-service/internal/domain/slot"
)

// SlotReader đọc 1 slot (không khoá); repository postgres triển khai interface này.
type SlotReader interface {
	GetSlotByID(ctx context.Context, slotID string) (*slotdomain.ExpertSlot, error)
}

// BookingConfirmationUsecase dựng dữ liệu cho trang xác nhận đặt lịch.
type BookingConfirmationUsecase interface {
	GetBookingConfirmation(ctx context.Context, query BookingConfirmationQuery) (*BookingConfirmation, error)
}

type BookingConfirmationQuery struct {
	PatientID        string
	ExpertID         string
	SlotID           string
	PatientRecordID  string
	SpecializationID string
}

type BookingConfirmation struct {
	Slot           ConfirmationSlot    `json:"slot"`
	Expert         ExpertInfo          `json:"expert"`
	PatientRecord  PatientRecordInfo   `json:"patient_record"`
	Specialization *SpecializationInfo `json:"specialization"`
}

type ConfirmationSlot struct {
	SlotID          string  `json:"slot_id"`
	StartTime       int64   `json:"start_time"`
	EndTime         int64   `json:"end_time"`
	Price           float64 `json:"price"`
	LockedExpiresAt int64   `json:"locked_expires_at"`
}

func (u *appointmentUsecase) GetBookingConfirmation(ctx context.Context, query BookingConfirmationQuery) (*BookingConfirmation, error) {
	reader, ok := u.repo.(SlotReader)
	if !ok {
		return nil, ErrBookingConfirmationUnavailable
	}

	slot, err := reader.GetSlotByID(ctx, query.SlotID)
	if err != nil {
		return nil, err
	}
	// Trang xác nhận chỉ hiển thị khi slot đang được chính bệnh nhân giữ chỗ.
	probe := &appointmentdomain.Appointment{PatientID: query.PatientID, ExpertID: query.ExpertID}
	if err := appointmentdomain.ValidateAppointmentCreationSlot(*slot, probe, time.Now().UnixMilli()); err != nil {
		if errors.Is(err, appointmentdomain.ErrAppointmentCreationSlotExpertMismatch) {
			return nil, ErrBookingProfileInvalid
		}
		return nil, ErrBookingSlotNotHeld
	}

	info, err := u.getBookingInfo(ctx, BookingInfoQuery{
		ExpertID:         query.ExpertID,
		PatientRecordID:  query.PatientRecordID,
		OwnerID:          query.PatientID,
		SpecializationID: query.SpecializationID,
	})
	if err != nil {
		return nil, err
	}

	return &BookingConfirmation{
		Slot: ConfirmationSlot{
			SlotID:          slot.SlotID,
			StartTime:       slot.StartTime,
			EndTime:         slot.EndTime,
			Price:           slot.Price,
			LockedExpiresAt: *slot.LockedExpiresAt,
		},
		Expert:         info.Expert,
		PatientRecord:  info.PatientRecord,
		Specialization: info.Specialization,
	}, nil
}

func (u *appointmentUsecase) getBookingInfo(ctx context.Context, query BookingInfoQuery) (*BookingInfo, error) {
	if u.profiles == nil {
		return nil, ErrProfileServiceUnavailable
	}
	return u.profiles.GetBookingInfo(ctx, query)
}
