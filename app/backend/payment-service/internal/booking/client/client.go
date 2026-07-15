// Package client định nghĩa port (interface) để Payment Service giao tiếp
// với Booking Service nội bộ.
//
// Thiết kế theo Hexagonal Architecture:
//   - usecase.go chỉ phụ thuộc vào interface BookingServiceClient này.
//   - Hiện tại: rest_client.go implement bằng HTTP REST.
//   - Tương lai: grpc_client.go implement bằng gRPC — chỉ cần thêm file, đổi 1 dòng main.go.
package client

import "context"

// BookingServiceClient là port để giao tiếp với Booking Service.
// Mọi implementation (REST, gRPC...) đều phải implement interface này.
type BookingServiceClient interface {
	// ConfirmAppointment thông báo cho Booking Service rằng thanh toán thành công.
	// Booking Service sẽ chuyển appointment → CONFIRMED, slot → OCCUPIED.
	ConfirmAppointment(ctx context.Context, appointmentID string) error

	// FailAppointment thông báo cho Booking Service rằng thanh toán thất bại.
	// Booking Service sẽ chuyển appointment → CANCELLED, slot → AVAILABLE.
	FailAppointment(ctx context.Context, appointmentID string) error
}
