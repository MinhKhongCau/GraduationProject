// File: internal/infrastructure/grpc/server.go
package grpc

import (
	"context"
	"log"
	"net"
	"strings"

	paymentdomain "payment-service/internal/domain/payment"
	"payment-service/internal/infrastructure/grpc/paymentpb"

	"github.com/google/uuid"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// BookingServiceCallerID là client_id duy nhất được phép gọi PaymentQueryService.
const BookingServiceCallerID = "booking-service"

// maxAppointmentIDs giới hạn số lịch hẹn mỗi lần tra cứu hàng loạt.
const maxAppointmentIDs = 100

// AppointmentPaymentReader đọc order đại diện (SUCCESS, hoặc mới nhất) của từng lịch hẹn.
type AppointmentPaymentReader interface {
	ListAppointmentPayments(ctx context.Context, appointmentIDs []uuid.UUID) ([]paymentdomain.PaymentOrder, error)
}

// CallerVerifier verify giá trị metadata "authorization" và trả về client_id của service gọi.
type CallerVerifier func(authorization string) (string, error)

// PaymentQueryServer cho booking-service tra số tiền thanh toán của lịch hẹn qua gRPC.
type PaymentQueryServer struct {
	paymentpb.UnimplementedPaymentQueryServiceServer

	payments AppointmentPaymentReader
}

func NewPaymentQueryServer(payments AppointmentPaymentReader) *PaymentQueryServer {
	return &PaymentQueryServer{payments: payments}
}

// Start lắng nghe gRPC trên addr (VD ":9004") trong goroutine riêng.
func Start(addr string, server *PaymentQueryServer, verify CallerVerifier) (*grpclib.Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpclib.NewServer(grpclib.UnaryInterceptor(InternalAuthInterceptor(verify)))
	paymentpb.RegisterPaymentQueryServiceServer(grpcServer, server)
	healthpb.RegisterHealthServer(grpcServer, health.NewServer())

	go func() {
		log.Printf("🛰️  gRPC server đang chạy tại %s", addr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("❌ gRPC server dừng: %v", err)
		}
	}()
	return grpcServer, nil
}

// InternalAuthInterceptor bắt buộc internal JWT của booking-service cho PaymentQueryService
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
		if callerID != BookingServiceCallerID {
			return nil, status.Errorf(codes.PermissionDenied, "caller %q is not allowed", callerID)
		}
		return handler(ctx, req)
	}
}

// GetAppointmentPayments trả số tiền thanh toán của nhiều lịch hẹn; lịch hẹn chưa có order bị bỏ qua.
func (s *PaymentQueryServer) GetAppointmentPayments(ctx context.Context, req *paymentpb.GetAppointmentPaymentsRequest) (*paymentpb.GetAppointmentPaymentsResponse, error) {
	raw := req.GetAppointmentIds()
	if len(raw) > maxAppointmentIDs {
		return nil, status.Errorf(codes.InvalidArgument, "tối đa %d appointment_id mỗi lần", maxAppointmentIDs)
	}
	ids := make([]uuid.UUID, 0, len(raw))
	for _, value := range raw {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "appointment_id không hợp lệ: %q", value)
		}
		ids = append(ids, id)
	}
	orders, err := s.payments.ListAppointmentPayments(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "đọc payment order: %v", err)
	}
	resp := &paymentpb.GetAppointmentPaymentsResponse{Payments: make([]*paymentpb.AppointmentPayment, 0, len(orders))}
	for i := range orders {
		order := &orders[i]
		if order.AppointmentID == nil {
			continue
		}
		payment := &paymentpb.AppointmentPayment{
			AppointmentId: order.AppointmentID.String(),
			OrderId:       order.ID.String(),
			AmountVnd:     order.GrossAmount.Int64(),
			Status:        toProtoOrderStatus(order.Status),
			Gateway:       order.Gateway,
		}
		if order.PaidAt != nil {
			payment.PaidAt = *order.PaidAt
		}
		resp.Payments = append(resp.Payments, payment)
	}
	return resp, nil
}

func toProtoOrderStatus(s paymentdomain.PaymentOrderStatus) paymentpb.OrderStatus {
	switch s {
	case paymentdomain.OrderStatusPending:
		return paymentpb.OrderStatus_ORDER_STATUS_PENDING
	case paymentdomain.OrderStatusSuccess:
		return paymentpb.OrderStatus_ORDER_STATUS_SUCCESS
	case paymentdomain.OrderStatusFailed:
		return paymentpb.OrderStatus_ORDER_STATUS_FAILED
	case paymentdomain.OrderStatusExpired:
		return paymentpb.OrderStatus_ORDER_STATUS_EXPIRED
	default:
		return paymentpb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
