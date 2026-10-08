package grpc

import (
	"context"
	"testing"

	"profile-service/internal/infrastructure/grpc/profilepb"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetBookingInfoRejectsInvalidIDs(t *testing.T) {
	valid := uuid.NewString()
	tests := map[string]*profilepb.GetBookingInfoRequest{
		"expert_id":         {ExpertId: "not-a-uuid", PatientRecordId: valid, OwnerId: valid},
		"patient_record_id": {ExpertId: valid, PatientRecordId: "", OwnerId: valid},
		"owner_id":          {ExpertId: valid, PatientRecordId: valid, OwnerId: "x"},
	}

	server := NewProfileQueryServer(nil, nil)
	for field, req := range tests {
		t.Run(field, func(t *testing.T) {
			_, err := server.GetBookingInfo(context.Background(), req)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument for bad %s, got %v", field, err)
			}
		})
	}
}

func TestListManagedExpertIdsRejectsInvalidAdminID(t *testing.T) {
	server := NewProfileQueryServer(nil, nil)
	_, err := server.ListManagedExpertIds(context.Background(), &profilepb.ListManagedExpertIdsRequest{AdminId: "x"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestGetProfileSummariesValidatesInput(t *testing.T) {
	server := NewProfileQueryServer(nil, nil)
	if _, err := server.GetProfileSummaries(context.Background(), &profilepb.GetProfileSummariesRequest{AuthIds: []string{"x"}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for bad id, got %v", err)
	}
	tooMany := make([]string, maxSummaryIDs+1)
	for i := range tooMany {
		tooMany[i] = uuid.NewString()
	}
	if _, err := server.GetProfileSummaries(context.Background(), &profilepb.GetProfileSummariesRequest{AuthIds: tooMany}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for too many ids, got %v", err)
	}
	resp, err := server.GetProfileSummaries(context.Background(), &profilepb.GetProfileSummariesRequest{})
	if err != nil || len(resp.GetProfiles()) != 0 {
		t.Fatalf("empty request should return empty list, got %v %v", resp, err)
	}
}
