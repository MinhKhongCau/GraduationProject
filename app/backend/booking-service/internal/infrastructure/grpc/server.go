// File: internal/infrastructure/grpc/server.go
package grpc

import (
	"context"
	"errors"
	"log"
	"net"
	"strings"

	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"booking-service/internal/infrastructure/grpc/bookingpb"

	"github.com/google/uuid"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// PaymentServiceCallerID là client_id duy nhất được phép gọi BookingPaymentService.
const PaymentServiceCallerID = "payment-service"

// PaymentResultHandler là phần usecase lịch hẹn mà gRPC server cần.
type PaymentResultHandler interface {
	HandlePaymentResult(command appappointment.HandlePaymentResultCommand) error
	GetAppointmentByID(appointmentID string) (*appointmentdomain.Appointment, error)
	GetAppointmentSummaries(ctx context.Context, appointmentIDs []string) ([]appointmentdomain.Appointment, error)
}

// maxSummaryIDs giới hạn số lịch hẹn mỗi lần tra cứu hàng loạt.
const maxSummaryIDs = 100

// CallerVerifier verify giá trị metadata "authorization" và trả về client_id của service gọi.
type CallerVerifier func(authorization string) (string, error)

// BookingPaymentServer nhận kết quả thanh toán từ payment-service qua gRPC.
type BookingPaymentServer struct {
	bookingpb.UnimplementedBookingPaymentServiceServer

	appointments PaymentResultHandler
}

func NewBookingPaymentServer(appointments PaymentResultHandler) *BookingPaymentServer {
	return &BookingPaymentServer{appointments: appointments}
}

// Start lắng nghe gRPC trên addr (VD ":9003") trong goroutine riêng.
func Start(addr string, server *BookingPaymentServer, verify CallerVerifier) (*grpclib.Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpclib.NewServer(grpclib.UnaryInterceptor(InternalAuthInterceptor(verify)))
	bookingpb.RegisterBookingPaymentServiceServer(grpcServer, server)
	healthpb.RegisterHealthServer(grpcServer, health.NewServer())

	go func() {
		log.Printf("🛰️  gRPC server đang chạy tại %s", addr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("❌ gRPC server dừng: %v", err)
		}
	}()
	return grpcServer, nil
}

// InternalAuthInterceptor bắt buộc internal JWT của payment-service cho BookingPaymentService
// (health check không cần token).
func InternalAuthInterceptor(verify CallerVerifier) grpclib.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpclib.UnaryServerInfo, handler grpclib.UnaryHandler) (interface{}, error) {
		if strings.HasPrefix(info.FullMethod, "/"+healthpb.Health_ServiceDesc.ServiceName+"/") {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization metadata")
		}
		callerID, err := verify(values[0])
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid internal token: %v", err)
		}
		if callerID != PaymentServiceCallerID {
			return nil, status.Errorf(codes.PermissionDenied, "caller %q is not allowed", callerID)
		}
		return handler(ctx, req)
	}
}

func (s *BookingPaymentServer) ApplyPaymentResult(ctx context.Context, req *bookingpb.ApplyPaymentResultRequest) (*bookingpb.ApplyPaymentResultResponse, error) {
	appointmentID := req.GetAppointmentId()
	if _, err := uuid.Parse(appointmentID); err != nil {
		return nil, status.Error(codes.InvalidArgument, "appointment_id không hợp lệ")
	}

	var (
		result    appappointment.PaymentResultStatus
		newStatus bookingpb.AppointmentStatus
	)
	switch req.GetResult() {
	case bookingpb.PaymentResult_PAYMENT_RESULT_SUCCESS:
		result, newStatus = appappointment.PaymentResultSuccess, bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED
	case bookingpb.PaymentResult_PAYMENT_RESULT_FAILED:
		result, newStatus = appappointment.PaymentResultFailed, bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED
	default:
		return nil, status.Error(codes.InvalidArgument, "result phải là SUCCESS hoặc FAILED")
	}

	err := s.appointments.HandlePaymentResult(appappointment.HandlePaymentResultCommand{
		AppointmentID: appointmentID,
		Status:        result,
	})
	if err != nil {
		log.Printf("⚠️  gRPC ApplyPaymentResult appointment_id=%s order_id=%s result=%s failed: %v",
			appointmentID, req.GetOrderId(), result, err)
		return nil, mapPaymentResultError(err)
	}
	log.Printf("✅ gRPC ApplyPaymentResult appointment_id=%s order_id=%s result=%s -> %s",
		appointmentID, req.GetOrderId(), result, newStatus)
	return &bookingpb.ApplyPaymentResultResponse{AppointmentId: appointmentID, Status: newStatus}, nil
}

func (s *BookingPaymentServer) GetAppointmentStatus(ctx context.Context, req *bookingpb.GetAppointmentStatusRequest) (*bookingpb.GetAppointmentStatusResponse, error) {
	appointmentID := req.GetAppointmentId()
	if _, err := uuid.Parse(appointmentID); err != nil {
		return nil, status.Error(codes.InvalidArgument, "appointment_id không hợp lệ")
	}
	appt, err := s.appointments.GetAppointmentByID(appointmentID)
	if err != nil {
		if errors.Is(err, appappointment.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "không tìm thấy lịch hẹn")
		}
		return nil, status.Errorf(codes.Internal, "truy vấn lịch hẹn: %v", err)
	}
	return &bookingpb.GetAppointmentStatusResponse{
		AppointmentId: appointmentID,
		Status:        toProtoStatus(appt.Status),
	}, nil
}

func mapPaymentResultError(err error) error {
	switch {
	case errors.Is(err, appappointment.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, appappointment.ErrInvalidPaymentResultStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, appappointment.ErrInvalidStatus), errors.Is(err, appappointment.ErrPaymentResultConflict):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProtoStatus(s appointmentdomain.AppointmentStatus) bookingpb.AppointmentStatus {
	switch s {
	case appointmentdomain.AppointmentStatusPendingPayment:
		return bookingpb.AppointmentStatus_APPOINTMENT_STATUS_PENDING_PAYMENT
	case appointmentdomain.AppointmentStatusConfirmed:
		return bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED
	case appointmentdomain.AppointmentStatusCancelled:
		return bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED
	case appointmentdomain.AppointmentStatusCompleted:
		return bookingpb.AppointmentStatus_APPOINTMENT_STATUS_COMPLETED
	default:
		return bookingpb.AppointmentStatus_APPOINTMENT_STATUS_UNSPECIFIED
	}
}

// GetAppointmentSummaries trả thông tin hiển thị của nhiều lịch hẹn; id không tồn tại bị bỏ qua.
func (s *BookingPaymentServer) GetAppointmentSummaries(ctx context.Context, req *bookingpb.GetAppointmentSummariesRequest) (*bookingpb.GetAppointmentSummariesResponse, error) {
	ids := req.GetAppointmentIds()
	if len(ids) > maxSummaryIDs {
		return nil, status.Errorf(codes.InvalidArgument, "tối đa %d appointment_id mỗi lần", maxSummaryIDs)
	}
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "appointment_id không hợp lệ: %q", id)
		}
	}
	appointments, err := s.appointments.GetAppointmentSummaries(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "đọc lịch hẹn: %v", err)
	}
	resp := &bookingpb.GetAppointmentSummariesResponse{Appointments: make([]*bookingpb.AppointmentSummary, 0, len(appointments))}
	for i := range appointments {
		a := &appointments[i]
		resp.Appointments = append(resp.Appointments, &bookingpb.AppointmentSummary{
			AppointmentId:      a.AppointmentID,
			PatientId:          a.PatientID,
			ExpertId:           a.ExpertID,
			Status:             toProtoStatus(a.Status),
			StartTime:          a.StartTime,
			EndTime:            a.EndTime,
			PriceVnd:           int64(a.Price),
			SpecializationName: a.SpecializationName,
			PatientFullName:    a.Patient.FullName,
		})
	}
	return resp, nil
}
