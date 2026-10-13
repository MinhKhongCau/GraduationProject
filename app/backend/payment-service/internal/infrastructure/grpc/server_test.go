package grpc

import (
	"context"
	"errors"
	"testing"

	paymentdomain "payment-service/internal/domain/payment"
	"payment-service/internal/infrastructure/grpc/paymentpb"

	"github.com/google/uuid"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakePaymentReader struct {
	orders []paymentdomain.PaymentOrder
	asked  []uuid.UUID
}

func (f *fakePaymentReader) ListAppointmentPayments(_ context.Context, ids []uuid.UUID) ([]paymentdomain.PaymentOrder, error) {
	f.asked = ids
	return f.orders, nil
}

func TestGetAppointmentPaymentsMapsOrders(t *testing.T) {
	appointmentID := uuid.New()
	orderID := uuid.New()
	paidAt := int64(1_700_000_000_000)
	reader := &fakePaymentReader{orders: []paymentdomain.PaymentOrder{
		{ID: orderID, AppointmentID: &appointmentID, GrossAmount: 350000, Status: paymentdomain.OrderStatusSuccess, Gateway: "VNPAY", PaidAt: &paidAt},
		{ID: uuid.New(), GrossAmount: 100000}, // top-up order without appointment is skipped
	}}
	resp, err := NewPaymentQueryServer(reader).GetAppointmentPayments(context.Background(),
		&paymentpb.GetAppointmentPaymentsRequest{AppointmentIds: []string{appointmentID.String()}})
	if err != nil {
		t.Fatal(err)
	}
	if len(reader.asked) != 1 || reader.asked[0] != appointmentID {
		t.Fatalf("unexpected lookup ids: %v", reader.asked)
	}
	if len(resp.GetPayments()) != 1 {
		t.Fatalf("expected one payment, got %v", resp.GetPayments())
	}
	got := resp.GetPayments()[0]
	if got.GetAppointmentId() != appointmentID.String() || got.GetOrderId() != orderID.String() || got.GetAmountVnd() != 350000 ||
		got.GetStatus() != paymentpb.OrderStatus_ORDER_STATUS_SUCCESS || got.GetGateway() != "VNPAY" || got.GetPaidAt() != paidAt {
		t.Fatalf("unexpected payment: %v", got)
	}
}

func TestGetAppointmentPaymentsRejectsBadInput(t *testing.T) {
	server := NewPaymentQueryServer(&fakePaymentReader{})
	_, err := server.GetAppointmentPayments(context.Background(), &paymentpb.GetAppointmentPaymentsRequest{AppointmentIds: []string{"not-a-uuid"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for bad id, got %v", err)
	}
	tooMany := make([]string, maxAppointmentIDs+1)
	for i := range tooMany {
		tooMany[i] = uuid.NewString()
	}
	_, err = server.GetAppointmentPayments(context.Background(), &paymentpb.GetAppointmentPaymentsRequest{AppointmentIds: tooMany})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for too many ids, got %v", err)
	}
}

func TestInternalAuthInterceptorOnlyAllowsBookingService(t *testing.T) {
	verify := func(authorization string) (string, error) {
		switch authorization {
		case "Bearer booking":
			return BookingServiceCallerID, nil
		case "Bearer other":
			return "chatroom-service", nil
		default:
			return "", errors.New("bad token")
		}
	}
	interceptor := InternalAuthInterceptor(verify)
	info := &grpclib.UnaryServerInfo{FullMethod: paymentpb.PaymentQueryService_GetAppointmentPayments_FullMethodName}
	handler := func(context.Context, interface{}) (interface{}, error) { return "ok", nil }

	cases := map[string]codes.Code{"": codes.Unauthenticated, "Bearer bad": codes.Unauthenticated, "Bearer other": codes.PermissionDenied, "Bearer booking": codes.OK}
	for authorization, want := range cases {
		ctx := context.Background()
		if authorization != "" {
			ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", authorization))
		}
		_, err := interceptor(ctx, nil, info, handler)
		if status.Code(err) != want {
			t.Fatalf("authorization %q: expected %s, got %v", authorization, want, err)
		}
	}
}
