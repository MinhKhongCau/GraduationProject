// File: internal/infrastructure/persistence/repository/patient_record_repository.go
package repository

import (
	"context"
	"errors"

	"profile-service/internal/infrastructure/persistence/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrPatientRecordLimit: tài khoản đã đạt số hồ sơ người khám tối đa.
var ErrPatientRecordLimit = errors.New("patient record limit reached")

// PatientRecordRepository đọc/ghi hồ sơ người khám, dùng chung cho HTTP handler và gRPC server.
type PatientRecordRepository struct {
	db *gorm.DB
}

func NewPatientRecordRepository(db *gorm.DB) *PatientRecordRepository {
	return &PatientRecordRepository{db: db}
}

// ListByOwner trả về hồ sơ của một tài khoản: hồ sơ SELF trước, sau đó theo thứ tự tạo.
func (r *PatientRecordRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.PatientRecord, error) {
	if err := r.syncSelfRecord(ctx, ownerID); err != nil {
		return nil, err
	}

	var records []models.PatientRecord
	err := r.db.WithContext(ctx).
		Where("owner_auth_id = ?", ownerID).
		Order(clause.OrderBy{Expression: clause.Expr{
			SQL:                "CASE WHEN relationship = ? THEN 0 ELSE 1 END, created_at ASC",
			Vars:               []any{models.RelationshipSelf},
			WithoutParentheses: true,
		}}).
		Find(&records).Error
	return records, err
}

// FindOwned tìm hồ sơ theo id, chỉ khi hồ sơ thuộc ownerID; ngược lại trả gorm.ErrRecordNotFound.
func (r *PatientRecordRepository) FindOwned(ctx context.Context, recordID, ownerID uuid.UUID) (*models.PatientRecord, error) {
	if err := r.syncSelfRecord(ctx, ownerID); err != nil {
		return nil, err
	}

	var record models.PatientRecord
	err := r.db.WithContext(ctx).
		Where("record_id = ? AND owner_auth_id = ?", recordID, ownerID).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Create thêm hồ sơ người thân cho ownerID, giới hạn MaxPatientRecordsPerOwner hồ sơ.
func (r *PatientRecordRepository) Create(ctx context.Context, record *models.PatientRecord) error {
	if err := r.syncSelfRecord(ctx, record.OwnerAuthID); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Khoá theo owner để 2 request song song không cùng vượt giới hạn.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", record.OwnerAuthID.String()).Error; err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&models.PatientRecord{}).Where("owner_auth_id = ?", record.OwnerAuthID).Count(&count).Error; err != nil {
			return err
		}
		if count >= models.MaxPatientRecordsPerOwner {
			return ErrPatientRecordLimit
		}
		return tx.Create(record).Error
	})
}

// syncSelfRecord tạo (hoặc cập nhật) hồ sơ SELF từ Profile của chính bệnh nhân, để hồ sơ này
// luôn khớp với thông tin họ sửa ở /profiles/me. Tài khoản không phải bệnh nhân thì bỏ qua.
func (r *PatientRecordRepository) syncSelfRecord(ctx context.Context, ownerID uuid.UUID) error {
	var profile models.Profile
	err := r.db.WithContext(ctx).
		Preload("PatientProfile").
		Where("auth_id = ? AND role = ?", ownerID, models.RolePatient).
		First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	record := models.NewSelfPatientRecord(&profile)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "owner_auth_id"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "relationship = 'SELF' AND deleted_at IS NULL"},
		}},
		DoUpdates: clause.AssignmentColumns([]string{
			"full_name", "date_of_birth", "gender", "phone_number", "email", "address", "updated_at",
		}),
	}).Create(&record).Error
}
