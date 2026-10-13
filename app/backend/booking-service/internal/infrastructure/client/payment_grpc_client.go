package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"booking-service/internal/infrastructure/grpc/paymentpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const paymentGRPCTimeout = 3 * time.Second

// paymentBatch khớp giới hạn số id mỗi lần của payment-service.
const paymentBatch = 100

// InternalTokenProvider cấp internal JWT (TokenManager) để gắn vào metadata gRPC.
type InternalTokenProvider interface {
	GetToken(ctx context.Context) (string, error)
	InvalidateToken()
}

// PaymentGRPCClient gọi payment-service qua gRPC (mạng Docker nội bộ, không TLS),
// gắn internal JWT của booking-service vào metadata; triển khai appappointment.PaymentDirectory.
type PaymentGRPCClient struct {
	conn   *grpc.ClientConn
	client paymentpb.PaymentQueryServiceClient
	tokens InternalTokenProvider
}

var _ appappointment.PaymentDirectory = (*PaymentGRPCClient)(nil)

// NewPaymentGRPCClient tạo kết nối lazy: không lỗi khi payment-service chưa sẵn sàng lúc khởi động.
func NewPaymentGRPCClient(addr string, tokens InternalTokenProvider) (*PaymentGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("init payment gRPC client %s: %w", addr, err)
	}
	return &PaymentGRPCClient{conn: conn, client: paymentpb.NewPaymentQueryServiceClient(conn), tokens: tokens}, nil
}

func (c *PaymentGRPCClient) Close() error {
	return c.conn.Close()
}

// GetAppointmentPayments trả số tiền thanh toán theo appointment id (chia lô 100);
// lịch hẹn chưa có order không có trong map.
func (c *PaymentGRPCClient) GetAppointmentPayments(ctx context.Context, appointmentIDs []string) (map[string]appointmentdomain.PaymentSummary, error) {
	result := make(map[string]appointmentdomain.PaymentSummary, len(appointmentIDs))
	for start := 0; start < len(appointmentIDs); start += paymentBatch {
		end := min(start+paymentBatch, len(appointmentIDs))
		resp, err := c.call(ctx, appointmentIDs[start:end])
		if status.Code(err) == codes.Unauthenticated {
			// Token có thể vừa hết hạn/bị thu hồi: lấy token mới và thử lại một lần.
			c.tokens.InvalidateToken()
			resp, err = c.call(ctx, appointmentIDs[start:end])
		}
		if err != nil {
			return result, fmt.Errorf("payment gRPC GetAppointmentPayments: %s", status.Convert(err).Message())
		}
		for _, p := range resp.GetPayments() {
			summary := appointmentdomain.PaymentSummary{
				OrderID:   p.GetOrderId(),
				AmountVND: p.GetAmountVnd(),
				Status:    strings.TrimPrefix(p.GetStatus().String(), "ORDER_STATUS_"),
				Gateway:   p.GetGateway(),
			}
			if paidAt := p.GetPaidAt(); paidAt > 0 {
				summary.PaidAt = &paidAt
			}
			result[p.GetAppointmentId()] = summary
		}
	}
	return result, nil
}

func (c *PaymentGRPCClient) call(ctx context.Context, ids []string) (*paymentpb.GetAppointmentPaymentsResponse, error) {
	token, err := c.tokens.GetToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "internal token unavailable: %v", err)
	}
	ctx, cancel := context.WithTimeout(metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token), paymentGRPCTimeout)
	defer cancel()
	return c.client.GetAppointmentPayments(ctx, &paymentpb.GetAppointmentPaymentsRequest{AppointmentIds: ids})
}
