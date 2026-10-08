package payment

import (
	"context"
	"log"
	paymentdomain "payment-service/internal/domain/payment"

	"github.com/google/uuid"
)

// PartySummary là thông tin hiển thị của bệnh nhân/chuyên gia (profile-service).
type PartySummary struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Email     string    `json:"email,omitempty"`
}

// AppointmentInfo là thông tin lịch hẹn của một đơn thanh toán (booking-service).
type AppointmentInfo struct {
	ID                 uuid.UUID `json:"id"`
	Status             string    `json:"status"`
	StartTime          int64     `json:"start_time"`
	EndTime            int64     `json:"end_time"`
	SpecializationName string    `json:"specialization_name,omitempty"`
	// Tên người khám trong hồ sơ lúc đặt lịch (có thể khác chủ tài khoản thanh toán).
	PatientFullName string `json:"patient_full_name,omitempty"`
}

// ProfileDirectory tra cứu hồ sơ hàng loạt qua profile-service (gRPC).
type ProfileDirectory interface {
	GetProfileSummaries(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]PartySummary, error)
}

// AppointmentDirectory tra cứu lịch hẹn hàng loạt qua booking-service (gRPC).
type AppointmentDirectory interface {
	GetAppointmentSummaries(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]AppointmentInfo, error)
}

// orderContext chứa thông tin ghép thêm cho một trang đơn thanh toán.
type orderContext struct {
	profiles     map[uuid.UUID]PartySummary
	appointments map[uuid.UUID]AppointmentInfo
}

func (oc orderContext) party(id uuid.UUID) *PartySummary {
	if p, ok := oc.profiles[id]; ok {
		return &p
	}
	return nil
}

func (oc orderContext) appointment(id *uuid.UUID) *AppointmentInfo {
	if id == nil {
		return nil
	}
	if a, ok := oc.appointments[*id]; ok {
		return &a
	}
	return nil
}

// loadOrderContext ghép tên bệnh nhân/chuyên gia và giờ khám bằng một lần gọi mỗi service.
// Lỗi chỉ được log: danh sách đơn vẫn trả về, chỉ thiếu thông tin ghép.
func (u *paymentUsecase) loadOrderContext(ctx context.Context, orders []paymentdomain.PaymentOrder) orderContext {
	var oc orderContext
	if len(orders) == 0 {
		return oc
	}
	parties := make([]uuid.UUID, 0, len(orders)*2)
	appointments := make([]uuid.UUID, 0, len(orders))
	seen := make(map[uuid.UUID]bool, len(orders)*3)
	add := func(list []uuid.UUID, id uuid.UUID) []uuid.UUID {
		if id == uuid.Nil || seen[id] {
			return list
		}
		seen[id] = true
		return append(list, id)
	}
	for i := range orders {
		parties = add(parties, orders[i].PayerID)
		parties = add(parties, orders[i].ExpertID)
		if orders[i].AppointmentID != nil {
			appointments = add(appointments, *orders[i].AppointmentID)
		}
	}
	if u.profiles != nil && len(parties) > 0 {
		profiles, err := u.profiles.GetProfileSummaries(ctx, parties)
		if err != nil {
			log.Printf("⚠️  payment order profile lookup failed: %v", err)
		}
		oc.profiles = profiles
	}
	if u.appointments != nil && len(appointments) > 0 {
		infos, err := u.appointments.GetAppointmentSummaries(ctx, appointments)
		if err != nil {
			log.Printf("⚠️  payment order appointment lookup failed: %v", err)
		}
		oc.appointments = infos
	}
	return oc
}
