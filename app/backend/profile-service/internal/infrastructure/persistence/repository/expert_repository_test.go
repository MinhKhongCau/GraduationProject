package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"profile-service/internal/application/expertprofile"
	"profile-service/internal/domain/expert"
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

// Response của PUT/PATCH chuyên gia trước đây là models.Profile được serialize trực tiếp.
// Test này khoá hợp đồng JSON: DTO mới phải serialize giống hệt model GORM cũ.
func TestExpertProfileView_MatchesLegacyJSONContract(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	profileID := uuid.New()
	legacy := models.Profile{
		ID:        profileID,
		Slug:      "dr-a-1234",
		Name:      "Dr. A",
		AuthID:    uuid.New(),
		Role:      models.RoleExpert,
		CreatedAt: now,
		UpdatedAt: now,
		ExpertProfile: &models.ExpertProfile{
			ProfileID:            profileID,
			PhoneNumber:          "0900",
			Email:                "a@example.com",
			AvatarURL:            "https://cdn/a.png",
			IntroductionVideoURL: "https://cdn/a.mp4",
			Bio:                  "bio",
			VerificationStatus:   "VERIFIED",
			Specializations: []models.Specialization{
				{SpecID: uuid.New(), Name: "Lo âu", Description: "desc", ImageURL: "img", IsActive: true},
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
		return expert.Reconstitute(expert.Snapshot{ProfileID: uuid.New(), AuthID: uuid.New(), Name: "Dr. A"})
	}

	t.Run("không đổi chuyên khoa thì không đụng bảng trung gian", func(t *testing.T) {
		db, mock := setupDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "profiles" SET "name"`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE "expert_profiles"`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		require.NoError(t, NewExpertRepository(db).Save(context.Background(), newExpert()))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("xoá hết chuyên khoa", func(t *testing.T) {
		db, mock := setupDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "profiles" SET "name"`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE "expert_profiles"`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`DELETE FROM "expert_specializations"`).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		e := newExpert()
		e.AssignSpecializations(nil)
		require.NoError(t, NewExpertRepository(db).Save(context.Background(), e))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
