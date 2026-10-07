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
