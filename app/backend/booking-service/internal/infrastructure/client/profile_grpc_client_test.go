package client

import (
	"context"
	"errors"
	"net"
	"testing"

	appappointment "booking-service/internal/application/appointment"
	"booking-service/internal/infrastructure/grpc/profilepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type stubProfileServer struct {
	profilepb.UnimplementedProfileQueryServiceServer
	err  error
	last *profilepb.GetBookingInfoRequest
}

func (s *stubProfileServer) GetBookingInfo(_ context.Context, req *profilepb.GetBookingInfoRequest) (*profilepb.GetBookingInfoResponse, error) {
	s.last = req
	if s.err != nil {
		return nil, s.err
	}
	spec := &profilepb.Specialization{SpecId: "spec-1", Code: "SPEC-001", Name: "Tâm lý"}
	return &profilepb.GetBookingInfoResponse{
		Expert: &profilepb.ExpertSummary{
			ExpertId:        req.GetExpertId(),
			FullName:        "Dr. A",
			Specializations: []*profilepb.Specialization{spec},
		},
		PatientRecord: &profilepb.PatientRecord{
			RecordId:     req.GetPatientRecordId(),
			OwnerId:      req.GetOwnerId(),
			FullName:     "Nguyen Van B",
			DateOfBirth:  "1990-01-02",
			Relationship: "SELF",
		},
		Specialization: spec,
	}, nil
}

func newBufconnClient(t *testing.T, server *stubProfileServer) *ProfileGRPCClient {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	profilepb.RegisterProfileQueryServiceServer(grpcServer, server)
	go grpcServer.Serve(listener)
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	client := &ProfileGRPCClient{conn: conn, client: profilepb.NewProfileQueryServiceClient(conn)}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestProfileGRPCClientMapsBookingInfo(t *testing.T) {
	server := &stubProfileServer{}
	client := newBufconnClient(t, server)

	info, err := client.GetBookingInfo(context.Background(), appappointment.BookingInfoQuery{
		ExpertID: "expert-1", PatientRecordID: "record-1", OwnerID: "patient-1", SpecializationID: "spec-1",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if server.last.GetOwnerId() != "patient-1" || server.last.GetSpecializationId() != "spec-1" {
		t.Fatalf("request not forwarded correctly: %v", server.last)
	}
	if info.Expert.FullName != "Dr. A" || len(info.Expert.Specializations) != 1 {
		t.Fatalf("unexpected expert %#v", info.Expert)
	}
	if info.PatientRecord.RecordID != "record-1" || info.PatientRecord.DateOfBirth != "1990-01-02" {
		t.Fatalf("unexpected patient record %#v", info.PatientRecord)
	}
	if info.Specialization == nil || info.Specialization.Name != "Tâm lý" {
		t.Fatalf("unexpected specialization %#v", info.Specialization)
	}
}

func TestProfileGRPCClientMapsStatusCodes(t *testing.T) {
	tests := map[codes.Code]error{
		codes.NotFound:           appappointment.ErrBookingProfileNotFound,
		codes.InvalidArgument:    appappointment.ErrBookingProfileInvalid,
		codes.FailedPrecondition: appappointment.ErrBookingSpecializationMismatch,
		codes.Unavailable:        appappointment.ErrProfileServiceUnavailable,
		codes.Internal:           appappointment.ErrProfileServiceUnavailable,
	}

	for code, want := range tests {
		t.Run(code.String(), func(t *testing.T) {
			client := newBufconnClient(t, &stubProfileServer{err: status.Error(code, "boom")})

			_, err := client.GetBookingInfo(context.Background(), appappointment.BookingInfoQuery{ExpertID: "e", PatientRecordID: "r", OwnerID: "o"})
			if !errors.Is(err, want) {
				t.Fatalf("expected %v, got %v", want, err)
			}
		})
	}
}
