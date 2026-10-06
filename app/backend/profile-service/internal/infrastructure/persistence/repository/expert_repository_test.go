package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"profile-service/internal/application/expertprofile"
	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/profile"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)
	return gormDB, mock
}

// Các API đọc profile serialize trực tiếp models.Profile, còn PUT/PATCH chuyên gia trả DTO.
// Test này khoá hợp đồng JSON: DTO phải serialize giống hệt model GORM.
func TestExpertProfileView_MatchesLegacyJSONContract(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	dob := time.Date(1990, 5, 6, 0, 0, 0, 0, time.UTC)
	profileID := uuid.New()
	legacy := models.Profile{
		ID:   profileID,
		Slug: "dr-a-1234",
		UserInformation: models.UserInformation{
			FullName:    "Dr. A",
			DateOfBirth: &dob,
			Gender:      "FEMALE",
			PhoneNumber: "0900",
			Country:     "Vietnam",
		},
		AuthID:    uuid.New(),
		Role:      models.RoleExpert,
		CreatedAt: now,
		UpdatedAt: now,
		ExpertProfile: &models.ExpertProfile{
			ProfileID:            profileID,
			Email:                "a@example.com",
			AvatarURL:            "https://cdn/a.png",
			IntroductionVideoURL: "https://cdn/a.mp4",
			Bio:                  "bio",
			VerificationStatus:   "VERIFIED",
			Specializations: []models.Specialization{
				{
					SpecID: uuid.New(), Code: "SPEC-001", Name: "Lo âu", Slug: "lo-au", Description: "desc",
					Symptoms: []string{"mất ngủ", "hồi hộp"}, Location: "Phòng 101", ImageURL: "img", IsActive: true,
				},
			},
		},
	}

	for name, specs := range map[string][]models.Specialization{
		"có chuyên khoa":    legacy.ExpertProfile.Specializations,
		"không chuyên khoa": nil,
	} {
		t.Run(name, func(t *testing.T) {
			row := legacy
			ep := *legacy.ExpertProfile
			ep.Specializations = specs
			row.ExpertProfile = &ep

			want, err := json.Marshal(row)
			require.NoError(t, err)
			got, err := json.Marshal(expertprofile.NewExpertProfileView(toExpert(&row)))
			require.NoError(t, err)

			assert.JSONEq(t, string(want), string(got))
		})
	}
}

func TestExpertRepository_FindByAuthID_NotFound(t *testing.T) {
	db, mock := setupDB(t)
	authID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE \(auth_id = \$1 AND role = \$2\)`).
		WithArgs(authID, string(models.RoleExpert), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := NewExpertRepository(db).FindByAuthID(context.Background(), authID)
	assert.ErrorIs(t, err, expert.ErrExpertNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExpertRepository_Save(t *testing.T) {
	newExpert := func() *expert.Expert {
		return expert.Reconstitute(expert.Snapshot{ProfileID: uuid.New(), AuthID: uuid.New(), UserInformation: profile.UserInformation{FullName: "Dr. A"}})
	}

	t.Run("không đổi chuyên khoa thì không đụng bảng trung gian", func(t *testing.T) {
		db, mock := setupDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "profiles" SET "country"=\$1,"date_of_birth"=\$2,"full_name"=\$3,"gender"=\$4,"phone_number"=\$5`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE "expert_profiles"`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		require.NoError(t, NewExpertRepository(db).Save(context.Background(), newExpert()))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("xoá hết chuyên khoa", func(t *testing.T) {
		db, mock := setupDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "profiles" SET "country"=\$1,"date_of_birth"=\$2,"full_name"=\$3,"gender"=\$4,"phone_number"=\$5`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE "expert_profiles"`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`DELETE FROM "expert_specializations"`).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		e := newExpert()
		e.AssignSpecializations(nil)
		require.NoError(t, NewExpertRepository(db).Save(context.Background(), e))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
