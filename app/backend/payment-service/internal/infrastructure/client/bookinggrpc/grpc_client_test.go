package bookinggrpc

import (
	"context"
	"errors"
	"net"
	"testing"

	apppayment "payment-service/internal/application/payment"
	"payment-service/internal/infrastructure/grpc/bookingpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type fakeTokens struct {
	token       string
	err         error
	invalidated int
}

func (f *fakeTokens) GetToken(context.Context) (string, error) { return f.token, f.err }
func (f *fakeTokens) InvalidateToken()                         { f.invalidated++ }

type fakeBookingServer struct {
	bookingpb.UnimplementedBookingPaymentServiceServer
	err      error
	requests []*bookingpb.ApplyPaymentResultRequest
	authz    []string
}

func (s *fakeBookingServer) ApplyPaymentResult(ctx context.Context, req *bookingpb.ApplyPaymentResultRequest) (*bookingpb.ApplyPaymentResultResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	s.authz = append(s.authz, md.Get("authorization")...)
	s.requests = append(s.requests, req)
	if s.err != nil {
		return nil, s.err
	}
	return &bookingpb.ApplyPaymentResultResponse{AppointmentId: req.GetAppointmentId(), Status: bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED}, nil
}

func (s *fakeBookingServer) GetAppointmentStatus(_ context.Context, req *bookingpb.GetAppointmentStatusRequest) (*bookingpb.GetAppointmentStatusResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &bookingpb.GetAppointmentStatusResponse{AppointmentId: req.GetAppointmentId(), Status: bookingpb.AppointmentStatus_APPOINTMENT_STATUS_PENDING_PAYMENT}, nil
}

func newTestClient(t *testing.T, server *fakeBookingServer, tokens *fakeTokens) *Client {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	grpcServer := grpc.NewServer()
	bookingpb.RegisterBookingPaymentServiceServer(grpcServer, server)
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return &Client{conn: conn, client: bookingpb.NewBookingPaymentServiceClient(conn), tokens: tokens}
}

func TestConfirmAndFailSendResultWithInternalToken(t *testing.T) {
	server := &fakeBookingServer{}
	client := newTestClient(t, server, &fakeTokens{token: "jwt-123"})

	if err := client.ConfirmAppointment(context.Background(), "appt-1"); err != nil {
		t.Fatalf("ConfirmAppointment: %v", err)
	}
	if err := client.FailAppointment(context.Background(), "appt-2"); err != nil {
		t.Fatalf("FailAppointment: %v", err)
	}
	if len(server.requests) != 2 ||
		server.requests[0].GetResult() != bookingpb.PaymentResult_PAYMENT_RESULT_SUCCESS || server.requests[0].GetAppointmentId() != "appt-1" ||
		server.requests[1].GetResult() != bookingpb.PaymentResult_PAYMENT_RESULT_FAILED || server.requests[1].GetAppointmentId() != "appt-2" {
		t.Fatalf("unexpected requests: %+v", server.requests)
	}
	for _, header := range server.authz {
		if header != "Bearer jwt-123" {
			t.Fatalf("unexpected authorization metadata %q", header)
		}
	}
}

func TestGetAppointmentStatusTrimsEnumPrefix(t *testing.T) {
	client := newTestClient(t, &fakeBookingServer{}, &fakeTokens{token: "t"})
	got, err := client.GetAppointmentStatus(context.Background(), "appt-1")
	if err != nil || got != "PENDING_PAYMENT" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestErrorMappingDrivesOutboxRetry(t *testing.T) {
	tests := []struct {
		code      codes.Code
		category  apppayment.BookingDeliveryFailureCategory
		retryable bool
	}{
		{codes.InvalidArgument, apppayment.BookingDeliveryBadRequest, false},
		{codes.NotFound, apppayment.BookingDeliveryNotFound, false},
		{codes.FailedPrecondition, apppayment.BookingDeliveryConflict, false},
		{codes.PermissionDenied, apppayment.BookingDeliveryAuthentication, false},
		{codes.Unauthenticated, apppayment.BookingDeliveryAuthentication, true},
		{codes.Unavailable, apppayment.BookingDeliveryNetwork, true},
		{codes.ResourceExhausted, apppayment.BookingDeliveryRateLimited, true},
		{codes.Internal, apppayment.BookingDeliveryUpstream, true},
	}
	for _, test := range tests {
		t.Run(test.code.String(), func(t *testing.T) {
			tokens := &fakeTokens{token: "t"}
			client := newTestClient(t, &fakeBookingServer{err: status.Error(test.code, "boom")}, tokens)

			err := client.ConfirmAppointment(context.Background(), "appt-1")
			var deliveryErr *apppayment.BookingDeliveryError
			if !errors.As(err, &deliveryErr) {
				t.Fatalf("expected BookingDeliveryError, got %v", err)
			}
			if deliveryErr.Category != test.category || deliveryErr.Retryable != test.retryable {
				t.Fatalf("got category=%s retryable=%v", deliveryErr.Category, deliveryErr.Retryable)
			}
			if (test.code == codes.Unauthenticated) != (tokens.invalidated == 1) {
				t.Fatalf("token invalidation mismatch: invalidated=%d", tokens.invalidated)
			}
		})
	}
}

func TestTokenFailureIsRetryable(t *testing.T) {
	server := &fakeBookingServer{}
	client := newTestClient(t, server, &fakeTokens{err: errors.New("auth-service down")})

	err := client.ConfirmAppointment(context.Background(), "appt-1")
	var deliveryErr *apppayment.BookingDeliveryError
	if !errors.As(err, &deliveryErr) || !deliveryErr.Retryable {
		t.Fatalf("expected retryable error, got %v", err)
	}
	if len(server.requests) != 0 {
		t.Fatal("request must not be sent without a token")
	}
}
