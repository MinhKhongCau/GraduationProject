// File: internal/infrastructure/grpc/server.go
package grpc

import (
	"context"
	"errors"
	"log"
	"net"

	"profile-service/internal/infrastructure/grpc/profilepb"
	"profile-service/internal/infrastructure/persistence/models"
	"profile-service/internal/infrastructure/persistence/repository"

	"github.com/google/uuid"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ProfileQueryServer phục vụ các truy vấn gRPC nội bộ (booking-service gọi vào).
type ProfileQueryServer struct {
	profilepb.UnimplementedProfileQueryServiceServer

	db      *gorm.DB
	records *repository.PatientRecordRepository
}

func NewProfileQueryServer(db *gorm.DB, records *repository.PatientRecordRepository) *ProfileQueryServer {
	return &ProfileQueryServer{db: db, records: records}
}

// Start lắng nghe gRPC trên addr (VD ":9002") trong goroutine riêng.
func Start(addr string, server *ProfileQueryServer) (*grpclib.Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpclib.NewServer()
	profilepb.RegisterProfileQueryServiceServer(grpcServer, server)
	healthpb.RegisterHealthServer(grpcServer, health.NewServer())

	go func() {
		log.Printf("🛰️  gRPC server đang chạy tại %s", addr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("❌ gRPC server dừng: %v", err)
		}
	}()
	return grpcServer, nil
}

func (s *ProfileQueryServer) GetBookingInfo(ctx context.Context, req *profilepb.GetBookingInfoRequest) (*profilepb.GetBookingInfoResponse, error) {
	expertID, err := uuid.Parse(req.GetExpertId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "expert_id không hợp lệ")
	}
	recordID, err := uuid.Parse(req.GetPatientRecordId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "patient_record_id không hợp lệ")
	}
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "owner_id không hợp lệ")
	}

	var expert models.Profile
	err = s.db.WithContext(ctx).
		Preload("ExpertProfile.Specializations").
		Where("auth_id = ? AND role = ?", expertID, models.RoleExpert).
		First(&expert).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "không tìm thấy chuyên gia")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "truy vấn chuyên gia: %v", err)
	}

	record, err := s.records.FindOwned(ctx, recordID, ownerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "không tìm thấy hồ sơ người khám")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "truy vấn hồ sơ người khám: %v", err)
	}

	resp := &profilepb.GetBookingInfoResponse{
		Expert:        toExpertSummary(&expert),
		PatientRecord: toPatientRecord(record),
	}

	if rawSpecID := req.GetSpecializationId(); rawSpecID != "" {
		specID, err := uuid.Parse(rawSpecID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "specialization_id không hợp lệ")
		}
		for _, spec := range resp.Expert.GetSpecializations() {
			if spec.GetSpecId() == specID.String() {
				resp.Specialization = spec
				break
			}
		}
		if resp.Specialization == nil {
			return nil, status.Error(codes.FailedPrecondition, "chuyên gia không phụ trách chuyên khoa đã chọn")
		}
	}

	return resp, nil
}

// ListManagedExpertIds trả về auth id các chuyên gia do Admin admin_id quản lý.
func (s *ProfileQueryServer) ListManagedExpertIds(ctx context.Context, req *profilepb.ListManagedExpertIdsRequest) (*profilepb.ListManagedExpertIdsResponse, error) {
	adminID, err := uuid.Parse(req.GetAdminId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "admin_id không hợp lệ")
	}

	var expertIDs []uuid.UUID
	err = s.db.WithContext(ctx).
		Model(&models.Profile{}).
		Joins("JOIN expert_profiles ON expert_profiles.profile_id = profiles.id").
		Where("profiles.role = ? AND expert_profiles.managed_by_admin_id = ?", models.RoleExpert, adminID).
		Pluck("profiles.auth_id", &expertIDs).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "truy vấn chuyên gia được quản lý: %v", err)
	}

	resp := &profilepb.ListManagedExpertIdsResponse{ExpertIds: make([]string, 0, len(expertIDs))}
	for _, id := range expertIDs {
		resp.ExpertIds = append(resp.ExpertIds, id.String())
	}
	return resp, nil
}

func toExpertSummary(p *models.Profile) *profilepb.ExpertSummary {
	summary := &profilepb.ExpertSummary{
		ExpertId: p.AuthID.String(),
		FullName: p.UserInformation.FullName,
	}
	if detail := p.ExpertProfile; detail != nil {
		summary.AvatarUrl = detail.AvatarURL
		summary.Email = detail.Email
		summary.VerificationStatus = detail.VerificationStatus
		for _, spec := range detail.Specializations {
			summary.Specializations = append(summary.Specializations, &profilepb.Specialization{
				SpecId: spec.SpecID.String(),
				Code:   spec.Code,
				Name:   spec.Name,
			})
		}
	}
	return summary
}

func toPatientRecord(r *models.PatientRecord) *profilepb.PatientRecord {
	record := &profilepb.PatientRecord{
		RecordId:     r.RecordID.String(),
		OwnerId:      r.OwnerAuthID.String(),
		FullName:     r.FullName,
		Gender:       r.Gender,
		PhoneNumber:  r.PhoneNumber,
		Email:        r.Email,
		Address:      r.Address,
		Relationship: string(r.Relationship),
	}
	if r.DateOfBirth != nil {
		record.DateOfBirth = r.DateOfBirth.Format("2006-01-02")
	}
	return record
}
