package expert

import (
	"testing"
	"time"

	"profile-service/internal/domain/profile"
	"profile-service/internal/domain/specialization"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newExpert(status string) *Expert {
	return Reconstitute(Snapshot{ProfileID: uuid.New(), AuthID: uuid.New(), UserInformation: profile.UserInformation{FullName: "Dr. A"}, VerificationStatus: status})
}

func TestReconstitute_DefaultsMissingStatusToUnverified(t *testing.T) {
	assert.Equal(t, StatusUnverified, newExpert("").VerificationStatus())
	assert.Equal(t, StatusVerified, newExpert("VERIFIED").VerificationStatus())
}

func TestChangeVerificationStatus(t *testing.T) {
	now := time.Now()
	adminID := uuid.New()

	t.Run("chỉ Admin được đổi trạng thái", func(t *testing.T) {
		e := newExpert("PENDING")
		err := e.ChangeVerificationStatus(StatusVerified, profile.RoleExpert, adminID, now)
		assert.ErrorIs(t, err, ErrVerificationRequiresAdmin)
		assert.Equal(t, StatusPending, e.VerificationStatus())
		assert.Empty(t, e.PullEvents())
	})

	t.Run("không phải Admin bị từ chối kể cả khi trạng thái không đổi", func(t *testing.T) {
		e := newExpert("PENDING")
		assert.ErrorIs(t, e.ChangeVerificationStatus(StatusPending, profile.RolePatient, adminID, now), ErrVerificationRequiresAdmin)
	})

	t.Run("trạng thái không hợp lệ", func(t *testing.T) {
		e := newExpert("PENDING")
		assert.ErrorIs(t, e.ChangeVerificationStatus("APPROVED", profile.RoleAdmin, adminID, now), ErrInvalidVerificationStatus)
	})

	t.Run("Admin đổi trạng thái phát sinh event", func(t *testing.T) {
		e := newExpert("PENDING")
		assert.NoError(t, e.ChangeVerificationStatus(StatusVerified, profile.RoleAdmin, adminID, now))
		assert.Equal(t, StatusVerified, e.VerificationStatus())

		events := e.PullEvents()
		assert.Len(t, events, 1)
		changed := events[0].(VerificationStatusChanged)
		assert.Equal(t, StatusPending, changed.From)
		assert.Equal(t, StatusVerified, changed.To)
		assert.Empty(t, e.PullEvents())
	})

	t.Run("giữ nguyên trạng thái thì không phát event", func(t *testing.T) {
		e := newExpert("VERIFIED")
		assert.NoError(t, e.ChangeVerificationStatus(StatusVerified, profile.RoleAdmin, adminID, now))
		assert.Empty(t, e.PullEvents())
	})
}

func TestChangeVerificationStatus_AssignsManager(t *testing.T) {
	now := time.Now()
	adminA, adminB := uuid.New(), uuid.New()

	t.Run("Admin duyệt trở thành người quản lý", func(t *testing.T) {
		e := newExpert("PENDING")
		assert.NoError(t, e.ChangeVerificationStatus(StatusVerified, profile.RoleAdmin, adminA, now))
		assert.Equal(t, &adminA, e.ManagerAdminID())
		assert.NotNil(t, e.Snapshot().VerifiedAt)

		changed := e.PullEvents()[0].(VerificationStatusChanged)
		assert.Equal(t, &adminA, changed.ManagerAdminID)
	})

	t.Run("Admin khác duyệt lại không đổi người quản lý", func(t *testing.T) {
		e := newExpert("PENDING")
		assert.NoError(t, e.ChangeVerificationStatus(StatusVerified, profile.RoleAdmin, adminA, now))
		assert.NoError(t, e.ChangeVerificationStatus(StatusRejected, profile.RoleAdmin, adminB, now))
		assert.Equal(t, &adminA, e.ManagerAdminID())
		assert.NoError(t, e.ChangeVerificationStatus(StatusVerified, profile.RoleAdmin, adminB, now))
		assert.Equal(t, &adminA, e.ManagerAdminID())
	})

	t.Run("từ chối không gán người quản lý", func(t *testing.T) {
		e := newExpert("PENDING")
		assert.NoError(t, e.ChangeVerificationStatus(StatusRejected, profile.RoleAdmin, adminA, now))
		assert.Nil(t, e.ManagerAdminID())
	})

	t.Run("chuyên gia VERIFIED chưa có người quản lý được nhận quản lý khi duyệt lại", func(t *testing.T) {
		e := newExpert("VERIFIED")
		assert.NoError(t, e.ChangeVerificationStatus(StatusVerified, profile.RoleAdmin, adminB, now))
		assert.Equal(t, &adminB, e.ManagerAdminID())
		assert.Empty(t, e.PullEvents())
	})

	t.Run("duyệt mà không xác định được Admin bị từ chối", func(t *testing.T) {
		e := newExpert("PENDING")
		assert.ErrorIs(t, e.ChangeVerificationStatus(StatusVerified, profile.RoleAdmin, uuid.Nil, now), ErrVerificationRequiresAdmin)
	})
}

func TestReviseDetails_KeepsVerificationStatus(t *testing.T) {
	e := newExpert("VERIFIED")
	e.ReviseDetails(Details{Bio: "new bio"})
	assert.Equal(t, "new bio", e.Details().Bio)
	assert.Equal(t, StatusVerified, e.VerificationStatus())
}

func TestAssignSpecializations_MarksChanged(t *testing.T) {
	e := newExpert("")
	assert.False(t, e.SpecializationsChanged())

	e.AssignSpecializations([]specialization.Specialization{{ID: uuid.New(), Name: "Tâm lý"}})
	assert.True(t, e.SpecializationsChanged())
	assert.Len(t, e.Specializations(), 1)

	e.AssignSpecializations(nil)
	assert.Empty(t, e.Specializations())
}
