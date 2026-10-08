package bookinggrpc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/infrastructure/client/httpclient"
	"payment-service/internal/infrastructure/grpc/bookingpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const callTimeout = 5 * time.Second

// Client gửi kết quả thanh toán sang booking-service qua gRPC (mạng Docker nội bộ, không TLS),
// gắn internal JWT vào metadata. GetPaymentEligibility vẫn đi qua REST client được truyền vào.
type Client struct {
	conn        *grpc.ClientConn
	client      bookingpb.BookingPaymentServiceClient
	tokens      httpclient.TokenProvider
	eligibility apppayment.BookingServiceClient
}

var (
	_ apppayment.BookingServiceClient = (*Client)(nil)
	_ apppayment.BookingStatusReader  = (*Client)(nil)
)

// New tạo kết nối lazy: không lỗi khi booking-service chưa sẵn sàng lúc khởi động.
func New(addr string, tokens httpclient.TokenProvider, eligibility apppayment.BookingServiceClient) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("init booking gRPC client %s: %w", addr, err)
	}
	return &Client{
		conn:        conn,
		client:      bookingpb.NewBookingPaymentServiceClient(conn),
		tokens:      tokens,
		eligibility: eligibility,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetPaymentEligibility(ctx context.Context, appointmentID string, payerID uuid.UUID) (*apppayment.PaymentEligibility, error) {
	return c.eligibility.GetPaymentEligibility(ctx, appointmentID, payerID)
}

func (c *Client) ConfirmAppointment(ctx context.Context, appointmentID string) error {
	return c.applyPaymentResult(ctx, appointmentID, bookingpb.PaymentResult_PAYMENT_RESULT_SUCCESS)
}

func (c *Client) FailAppointment(ctx context.Context, appointmentID string) error {
	return c.applyPaymentResult(ctx, appointmentID, bookingpb.PaymentResult_PAYMENT_RESULT_FAILED)
}

func (c *Client) applyPaymentResult(ctx context.Context, appointmentID string, result bookingpb.PaymentResult) error {
	ctx, cancel, err := c.outgoingContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()

	resp, err := c.client.ApplyPaymentResult(ctx, &bookingpb.ApplyPaymentResultRequest{
		AppointmentId: appointmentID,
		Result:        result,
	})
	if err != nil {
		return c.mapError("apply payment result", err)
	}
	log.Printf("[booking_grpc] appointment %s -> %s accepted by booking-service (status=%s)",
		appointmentID, result, resp.GetStatus())
	return nil
}

// GetAppointmentStatus trả về trạng thái lịch hẹn dạng PENDING_PAYMENT | CONFIRMED | CANCELLED | COMPLETED.
func (c *Client) GetAppointmentStatus(ctx context.Context, appointmentID string) (string, error) {
	ctx, cancel, err := c.outgoingContext(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()

	resp, err := c.client.GetAppointmentStatus(ctx, &bookingpb.GetAppointmentStatusRequest{AppointmentId: appointmentID})
	if err != nil {
		return "", c.mapError("get appointment status", err)
	}
	return strings.TrimPrefix(resp.GetStatus().String(), "APPOINTMENT_STATUS_"), nil
}

func (c *Client) outgoingContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	token, err := c.tokens.GetToken(ctx)
	if err != nil {
		return nil, nil, retryable(apppayment.BookingDeliveryAuthentication, "internal token unavailable: "+err.Error())
	}
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token), cancel, nil
}

func (c *Client) mapError(operation string, err error) error {
	message := fmt.Sprintf("booking gRPC %s: %s", operation, status.Convert(err).Message())
	switch status.Code(err) {
	case codes.InvalidArgument:
		return permanent(apppayment.BookingDeliveryBadRequest, message)
	case codes.NotFound:
		return permanent(apppayment.BookingDeliveryNotFound, message)
	case codes.FailedPrecondition, codes.AlreadyExists:
		return permanent(apppayment.BookingDeliveryConflict, message)
	case codes.Unauthenticated:
		// Token có thể vừa hết hạn/bị thu hồi: bỏ cache để lần retry xin token mới.
		c.tokens.InvalidateToken()
		return retryable(apppayment.BookingDeliveryAuthentication, message)
	case codes.PermissionDenied:
		return permanent(apppayment.BookingDeliveryAuthentication, message)
	case codes.DeadlineExceeded:
		return retryable(apppayment.BookingDeliveryTimeout, message)
	case codes.ResourceExhausted:
		return retryable(apppayment.BookingDeliveryRateLimited, message)
	case codes.Unavailable, codes.Canceled:
		return retryable(apppayment.BookingDeliveryNetwork, message)
	default:
		return retryable(apppayment.BookingDeliveryUpstream, message)
	}
}

func retryable(category apppayment.BookingDeliveryFailureCategory, message string) error {
	return &apppayment.BookingDeliveryError{Category: category, Retryable: true, Message: message}
}

func permanent(category apppayment.BookingDeliveryFailureCategory, message string) error {
	return &apppayment.BookingDeliveryError{Category: category, Retryable: false, Message: message}
}

var _ apppayment.AppointmentDirectory = (*Client)(nil)

// summaryBatch khớp giới hạn số id mỗi lần của booking-service.
const summaryBatch = 100

// GetAppointmentSummaries tra giờ khám, chuyên khoa, người khám của nhiều lịch hẹn (chia lô 100).
func (c *Client) GetAppointmentSummaries(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]apppayment.AppointmentInfo, error) {
	result := make(map[uuid.UUID]apppayment.AppointmentInfo, len(ids))
	for start := 0; start < len(ids); start += summaryBatch {
		end := min(start+summaryBatch, len(ids))
		raw := make([]string, 0, end-start)
		for _, id := range ids[start:end] {
			raw = append(raw, id.String())
		}
		callCtx, cancel, err := c.outgoingContext(ctx)
		if err != nil {
			return result, err
		}
		resp, err := c.client.GetAppointmentSummaries(callCtx, &bookingpb.GetAppointmentSummariesRequest{AppointmentIds: raw})
		cancel()
		if err != nil {
			return result, c.mapError("get appointment summaries", err)
		}
		for _, a := range resp.GetAppointments() {
			id, err := uuid.Parse(a.GetAppointmentId())
			if err != nil {
				continue
			}
			result[id] = apppayment.AppointmentInfo{
				ID:                 id,
				Status:             strings.TrimPrefix(a.GetStatus().String(), "APPOINTMENT_STATUS_"),
				StartTime:          a.GetStartTime(),
				EndTime:            a.GetEndTime(),
				SpecializationName: a.GetSpecializationName(),
				PatientFullName:    a.GetPatientFullName(),
			}
		}
	}
	return result, nil
}
