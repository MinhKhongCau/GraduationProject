// Package client định nghĩa port (interface) để Payment Service giao tiếp
// với Booking Service nội bộ.
//
// Thiết kế theo Hexagonal Architecture:
//   - usecase.go chỉ phụ thuộc vào interface BookingServiceClient này.
//   - Hiện tại: rest_client.go implement bằng HTTP REST.
//   - Tương lai: grpc_client.go implement bằng gRPC — chỉ cần thêm file, đổi 1 dòng main.go.
package client

import (
	"context"
	"errors"
)

var ErrAppointmentNotFound = errors.New("booking appointment not found")

type AppointmentStatus int

const (
	AppointmentStatusPendingPayment AppointmentStatus = 0
	AppointmentStatusConfirmed      AppointmentStatus = 1
	AppointmentStatusCancelled      AppointmentStatus = 2
)

func (s AppointmentStatus) String() string {
	switch s {
	case AppointmentStatusPendingPayment:
		return "PENDING_PAYMENT"
	case AppointmentStatusConfirmed:
		return "CONFIRMED"
	case AppointmentStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type Appointment struct {
	AppointmentID string            `json:"appointment_id"`
	PatientID     string            `json:"patient_id"`
	ExpertID      string            `json:"expert_id"`
	Status        AppointmentStatus `json:"status"`
	Price         float64           `json:"price"` // Giá thực tế của lịch khám
}

// BookingServiceClient là port để giao tiếp với Booking Service.
// Mọi implementation (REST, gRPC...) đều phải implement interface này.
type BookingServiceClient interface {
	// GetAppointment lấy thông tin lịch hẹn từ Booking Service
	GetAppointment(ctx context.Context, appointmentID string) (*Appointment, error)

	// ConfirmAppointment thông báo cho Booking Service rằng thanh toán thành công.
	// Booking Service sẽ chuyển appointment → CONFIRMED, slot → OCCUPIED.
	ConfirmAppointment(ctx context.Context, appointmentID string) error

	// FailAppointment thông báo cho Booking Service rằng thanh toán thất bại.
	// Booking Service sẽ chuyển appointment → CANCELLED, slot → AVAILABLE.
	FailAppointment(ctx context.Context, appointmentID string) error
}
