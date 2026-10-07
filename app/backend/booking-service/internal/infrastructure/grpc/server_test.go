package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"booking-service/internal/infrastructure/grpc/bookingpb"

	"github.com/google/uuid"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type fakeAppointments struct {
	commands []appappointment.HandlePaymentResultCommand
	applyErr error
	appt     *appointmentdomain.Appointment
	getErr   error
}

func (f *fakeAppointments) HandlePaymentResult(command appappointment.HandlePaymentResultCommand) error {
	f.commands = append(f.commands, command)
	return f.applyErr
}

func (f *fakeAppointments) GetAppointmentByID(string) (*appointmentdomain.Appointment, error) {
	return f.appt, f.getErr
}

// verifier chấp nhận "Bearer <client_id>" để test không cần JWT thật.
func fakeVerifier(header string) (string, error) {
	var callerID string
	if _, err := fmt.Sscanf(header, "Bearer %s", &callerID); err != nil {
		return "", errors.New("bad token")
	}
	return callerID, nil
}

func startTestServer(t *testing.T, appointments PaymentResultHandler) bookingpb.BookingPaymentServiceClient {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	server := grpclib.NewServer(grpclib.UnaryInterceptor(InternalAuthInterceptor(fakeVerifier)))
	bookingpb.RegisterBookingPaymentServiceServer(server, NewBookingPaymentServer(appointments))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	conn, err := grpclib.NewClient("passthrough:///bufnet",
		grpclib.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return bookingpb.NewBookingPaymentServiceClient(conn)
}

func asCaller(callerID string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+callerID)
}

func TestApplyPaymentResultMapsResultToUsecase(t *testing.T) {
	tests := []struct {
		result bookingpb.PaymentResult
		want   appappointment.PaymentResultStatus
		status bookingpb.AppointmentStatus
	}{
		{bookingpb.PaymentResult_PAYMENT_RESULT_SUCCESS, appappointment.PaymentResultSuccess, bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED},
		{bookingpb.PaymentResult_PAYMENT_RESULT_FAILED, appappointment.PaymentResultFailed, bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED},
	}
	for _, test := range tests {
		t.Run(test.result.String(), func(t *testing.T) {
			appointments := &fakeAppointments{}
			client := startTestServer(t, appointments)
			appointmentID := uuid.NewString()

			resp, err := client.ApplyPaymentResult(asCaller(PaymentServiceCallerID), &bookingpb.ApplyPaymentResultRequest{
				AppointmentId: appointmentID, OrderId: uuid.NewString(), Result: test.result,
			})
			if err != nil {
				t.Fatalf("ApplyPaymentResult: %v", err)
			}
			if resp.GetStatus() != test.status || resp.GetAppointmentId() != appointmentID {
				t.Fatalf("unexpected response: %+v", resp)
			}
			if len(appointments.commands) != 1 || appointments.commands[0].AppointmentID != appointmentID || appointments.commands[0].Status != test.want {
				t.Fatalf("unexpected commands: %+v", appointments.commands)
			}
		})
	}
}

func TestApplyPaymentResultErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{"not found", appappointment.ErrNotFound, codes.NotFound},
		{"conflict", fmt.Errorf("%w: slot", appappointment.ErrPaymentResultConflict), codes.FailedPrecondition},
		{"invalid status", appappointment.ErrInvalidPaymentResultStatus, codes.InvalidArgument},
		{"db down", errors.New("connection refused"), codes.Internal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := startTestServer(t, &fakeAppointments{applyErr: test.err})
			_, err := client.ApplyPaymentResult(asCaller(PaymentServiceCallerID), &bookingpb.ApplyPaymentResultRequest{
				AppointmentId: uuid.NewString(), Result: bookingpb.PaymentResult_PAYMENT_RESULT_SUCCESS,
			})
			if status.Code(err) != test.code {
				t.Fatalf("expected %s, got %v", test.code, err)
			}
		})
	}
}

func TestApplyPaymentResultValidatesRequest(t *testing.T) {
	appointments := &fakeAppointments{}
	client := startTestServer(t, appointments)

	_, err := client.ApplyPaymentResult(asCaller(PaymentServiceCallerID), &bookingpb.ApplyPaymentResultRequest{
		AppointmentId: "not-a-uuid", Result: bookingpb.PaymentResult_PAYMENT_RESULT_SUCCESS,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for bad id, got %v", err)
	}
	_, err = client.ApplyPaymentResult(asCaller(PaymentServiceCallerID), &bookingpb.ApplyPaymentResultRequest{
		AppointmentId: uuid.NewString(),
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for unspecified result, got %v", err)
	}
	if len(appointments.commands) != 0 {
		t.Fatalf("invalid requests must not reach the usecase: %+v", appointments.commands)
	}
}

func TestBookingPaymentServiceRequiresPaymentServiceToken(t *testing.T) {
	appointments := &fakeAppointments{}
	client := startTestServer(t, appointments)
	req := &bookingpb.ApplyPaymentResultRequest{AppointmentId: uuid.NewString(), Result: bookingpb.PaymentResult_PAYMENT_RESULT_SUCCESS}

	if _, err := client.ApplyPaymentResult(context.Background(), req); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated without token, got %v", err)
	}
	if _, err := client.ApplyPaymentResult(asCaller("chatbot-service"), req); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied for other caller, got %v", err)
	}
	if len(appointments.commands) != 0 {
		t.Fatalf("unauthorized calls must not reach the usecase: %+v", appointments.commands)
	}
}

func TestGetAppointmentStatus(t *testing.T) {
	appointmentID := uuid.NewString()
	client := startTestServer(t, &fakeAppointments{appt: &appointmentdomain.Appointment{
		AppointmentID: appointmentID, Status: appointmentdomain.AppointmentStatusConfirmed,
	}})

	resp, err := client.GetAppointmentStatus(asCaller(PaymentServiceCallerID), &bookingpb.GetAppointmentStatusRequest{AppointmentId: appointmentID})
	if err != nil {
		t.Fatalf("GetAppointmentStatus: %v", err)
	}
	if resp.GetStatus() != bookingpb.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED {
		t.Fatalf("unexpected status: %s", resp.GetStatus())
	}

	missing := startTestServer(t, &fakeAppointments{getErr: appappointment.ErrNotFound})
	if _, err := missing.GetAppointmentStatus(asCaller(PaymentServiceCallerID), &bookingpb.GetAppointmentStatusRequest{AppointmentId: appointmentID}); status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
