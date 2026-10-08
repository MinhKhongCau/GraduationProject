package client

import (
	"context"
	"fmt"
	"log"
	"time"

	appappointment "booking-service/internal/application/appointment"
	appointmentdomain "booking-service/internal/domain/appointment"
	"booking-service/internal/infrastructure/grpc/profilepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const profileGRPCTimeout = 3 * time.Second

// ProfileGRPCClient gọi profile-service qua gRPC (mạng Docker nội bộ, không TLS),
// triển khai appappointment.ProfileGateway.
type ProfileGRPCClient struct {
	conn   *grpc.ClientConn
	client profilepb.ProfileQueryServiceClient
}

var _ appappointment.ProfileGateway = (*ProfileGRPCClient)(nil)

// NewProfileGRPCClient tạo kết nối lazy: không lỗi khi profile-service chưa sẵn sàng lúc khởi động.
func NewProfileGRPCClient(addr string) (*ProfileGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("init profile gRPC client %s: %w", addr, err)
	}
	return &ProfileGRPCClient{conn: conn, client: profilepb.NewProfileQueryServiceClient(conn)}, nil
}

func (c *ProfileGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *ProfileGRPCClient) GetBookingInfo(ctx context.Context, query appappointment.BookingInfoQuery) (*appappointment.BookingInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, profileGRPCTimeout)
	defer cancel()

	resp, err := c.client.GetBookingInfo(ctx, &profilepb.GetBookingInfoRequest{
		ExpertId:         query.ExpertID,
		PatientRecordId:  query.PatientRecordID,
		OwnerId:          query.OwnerID,
		SpecializationId: query.SpecializationID,
	})
	if err != nil {
		return nil, mapProfileGRPCError(err)
	}
	return toBookingInfo(resp), nil
}

func mapProfileGRPCError(err error) error {
	switch status.Code(err) {
	case codes.NotFound:
		return appappointment.ErrBookingProfileNotFound
	case codes.InvalidArgument:
		return appappointment.ErrBookingProfileInvalid
	case codes.FailedPrecondition:
		return appappointment.ErrBookingSpecializationMismatch
	default:
		log.Printf("⚠️  profile-service gRPC GetBookingInfo failed: %v", err)
		return appappointment.ErrProfileServiceUnavailable
	}
}

func toBookingInfo(resp *profilepb.GetBookingInfoResponse) *appappointment.BookingInfo {
	expert := resp.GetExpert()
	record := resp.GetPatientRecord()

	info := &appappointment.BookingInfo{
		Expert: appappointment.ExpertInfo{
			ExpertID:           expert.GetExpertId(),
			FullName:           expert.GetFullName(),
			AvatarURL:          expert.GetAvatarUrl(),
			Email:              expert.GetEmail(),
			VerificationStatus: expert.GetVerificationStatus(),
			Specializations:    make([]appappointment.SpecializationInfo, 0, len(expert.GetSpecializations())),
		},
		PatientRecord: appappointment.PatientRecordInfo{
			RecordID:     record.GetRecordId(),
			FullName:     record.GetFullName(),
			DateOfBirth:  record.GetDateOfBirth(),
			Gender:       record.GetGender(),
			PhoneNumber:  record.GetPhoneNumber(),
			Email:        record.GetEmail(),
			Address:      record.GetAddress(),
			Relationship: record.GetRelationship(),
		},
	}
	for _, spec := range expert.GetSpecializations() {
		info.Expert.Specializations = append(info.Expert.Specializations, toSpecializationInfo(spec))
	}
	if spec := resp.GetSpecialization(); spec != nil {
		converted := toSpecializationInfo(spec)
		info.Specialization = &converted
	}
	return info
}

func toSpecializationInfo(spec *profilepb.Specialization) appappointment.SpecializationInfo {
	return appappointment.SpecializationInfo{
		SpecID: spec.GetSpecId(),
		Code:   spec.GetCode(),
		Name:   spec.GetName(),
	}
}

var _ appappointment.ProfileDirectory = (*ProfileGRPCClient)(nil)

// profileSummaryBatch khớp giới hạn số id mỗi lần của profile-service.
const profileSummaryBatch = 100

// GetProfileSummaries tra cứu hồ sơ hàng loạt, tự chia lô theo giới hạn của profile-service.
func (c *ProfileGRPCClient) GetProfileSummaries(ctx context.Context, authIDs []string) (map[string]appointmentdomain.ParticipantProfile, error) {
	result := make(map[string]appointmentdomain.ParticipantProfile, len(authIDs))
	for start := 0; start < len(authIDs); start += profileSummaryBatch {
		end := min(start+profileSummaryBatch, len(authIDs))
		callCtx, cancel := context.WithTimeout(ctx, profileGRPCTimeout)
		resp, err := c.client.GetProfileSummaries(callCtx, &profilepb.GetProfileSummariesRequest{AuthIds: authIDs[start:end]})
		cancel()
		if err != nil {
			return nil, fmt.Errorf("profile-service GetProfileSummaries: %w", err)
		}
		for _, p := range resp.GetProfiles() {
			result[p.GetAuthId()] = appointmentdomain.ParticipantProfile{
				AuthID:             p.GetAuthId(),
				Role:               p.GetRole(),
				FullName:           p.GetFullName(),
				AvatarURL:          p.GetAvatarUrl(),
				Email:              p.GetEmail(),
				PhoneNumber:        p.GetPhoneNumber(),
				VerificationStatus: p.GetVerificationStatus(),
			}
		}
	}
	return result, nil
}

// ListManagedExpertIDs trả auth id các chuyên gia do adminID quản lý (Admin đã duyệt).
func (c *ProfileGRPCClient) ListManagedExpertIDs(ctx context.Context, adminID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, profileGRPCTimeout)
	defer cancel()
	resp, err := c.client.ListManagedExpertIds(ctx, &profilepb.ListManagedExpertIdsRequest{AdminId: adminID})
	if err != nil {
		return nil, fmt.Errorf("profile-service ListManagedExpertIds: %w", err)
	}
	return resp.GetExpertIds(), nil
}
