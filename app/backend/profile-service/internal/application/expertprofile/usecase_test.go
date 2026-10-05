package expertprofile

import (
	"context"
	"errors"
	"testing"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/profile"
	"profile-service/internal/domain/specialization"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExpertRepo struct {
	experts map[uuid.UUID]expert.Snapshot
	saved   *expert.Expert
	saveErr error
}

func (r *fakeExpertRepo) FindByAuthID(_ context.Context, authID uuid.UUID) (*expert.Expert, error) {
	s, ok := r.experts[authID]
	if !ok {
		return nil, expert.ErrExpertNotFound
	}
	return expert.Reconstitute(s), nil
}

func (r *fakeExpertRepo) Save(_ context.Context, e *expert.Expert) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = e
	return nil
}

type fakeSpecRepo struct {
	specs map[string]specialization.Specialization
	calls int
}

func (r *fakeSpecRepo) FindByIDs(_ context.Context, ids []string) ([]specialization.Specialization, error) {
	r.calls++
	var found []specialization.Specialization
	for _, id := range ids {
		if spec, ok := r.specs[id]; ok {
			found = append(found, spec)
		}
	}
	return found, nil
}

type fakePublisher struct{ events []expert.Event }

func (p *fakePublisher) Publish(_ context.Context, events []expert.Event) {
	p.events = append(p.events, events...)
}

type fixture struct {
	authID    uuid.UUID
	spec      specialization.Specialization
	experts   *fakeExpertRepo
	specs     *fakeSpecRepo
	publisher *fakePublisher
}

func newFixture() *fixture {
	authID := uuid.New()
	existingSpec := specialization.Specialization{ID: uuid.New(), Name: "Trầm cảm"}
	newSpec := specialization.Specialization{ID: uuid.New(), Name: "Lo âu"}
	return &fixture{
		authID: authID,
		spec:   newSpec,
		experts: &fakeExpertRepo{experts: map[uuid.UUID]expert.Snapshot{
			authID: {
				ProfileID:          uuid.New(),
				AuthID:             authID,
				Slug:               "dr-a",
				Name:               "Dr. A",
				Details:            expert.Details{PhoneNumber: "0900", Email: "a@example.com", Bio: "old bio"},
				VerificationStatus: "VERIFIED",
				Specializations:    []specialization.Specialization{existingSpec},
			},
		}},
		specs:     &fakeSpecRepo{specs: map[string]specialization.Specialization{newSpec.ID.String(): newSpec}},
		publisher: &fakePublisher{},
	}
}

func strPtr(s string) *string { return &s }

func TestReplaceExpertProfile(t *testing.T) {
	t.Run("PUT thay toàn bộ thông tin nhưng giữ trạng thái xác minh", func(t *testing.T) {
		f := newFixture()
		uc := NewReplaceExpertProfile(f.experts, f.specs, f.publisher)

		view, err := uc.Execute(context.Background(), ReplaceCommand{AuthID: f.authID, Name: "Dr. B", Bio: "new bio"})
		require.NoError(t, err)

		assert.Equal(t, "Dr. B", view.Name)
		assert.Equal(t, profile.RoleExpert, view.Role)
		assert.Equal(t, "new bio", view.ExpertProfile.Bio)
		assert.Empty(t, view.ExpertProfile.PhoneNumber, "PUT xoá các field không gửi lên")
		assert.Equal(t, "VERIFIED", view.ExpertProfile.VerificationStatus)
		assert.Len(t, view.ExpertProfile.Specializations, 1, "specialization_ids nil = giữ nguyên")
		assert.Zero(t, f.specs.calls)
		assert.False(t, f.experts.saved.SpecializationsChanged())
	})

	t.Run("specialization_ids rỗng xoá hết, id không tồn tại bị bỏ qua", func(t *testing.T) {
		f := newFixture()
		uc := NewReplaceExpertProfile(f.experts, f.specs, f.publisher)

		view, err := uc.Execute(context.Background(), ReplaceCommand{AuthID: f.authID, Name: "Dr. A", SpecializationIDs: []string{}})
		require.NoError(t, err)
		assert.Empty(t, view.ExpertProfile.Specializations)
		assert.True(t, f.experts.saved.SpecializationsChanged())

		view, err = uc.Execute(context.Background(), ReplaceCommand{
			AuthID:            f.authID,
			Name:              "Dr. A",
			SpecializationIDs: []string{f.spec.ID.String(), uuid.NewString()},
		})
		require.NoError(t, err)
		require.Len(t, view.ExpertProfile.Specializations, 1)
		assert.Equal(t, f.spec.ID, view.ExpertProfile.Specializations[0].SpecID)
	})

	t.Run("không tìm thấy chuyên gia", func(t *testing.T) {
		f := newFixture()
		uc := NewReplaceExpertProfile(f.experts, f.specs, f.publisher)

		_, err := uc.Execute(context.Background(), ReplaceCommand{AuthID: uuid.New(), Name: "X"})
		assert.ErrorIs(t, err, expert.ErrExpertNotFound)
	})

	t.Run("lỗi lưu trữ được trả về", func(t *testing.T) {
		f := newFixture()
		f.experts.saveErr = errors.New("db down")
		uc := NewReplaceExpertProfile(f.experts, f.specs, f.publisher)

		_, err := uc.Execute(context.Background(), ReplaceCommand{AuthID: f.authID, Name: "X"})
		assert.EqualError(t, err, "db down")
	})
}

func TestPatchExpertProfile(t *testing.T) {
	t.Run("PATCH chỉ cập nhật field được gửi lên", func(t *testing.T) {
		f := newFixture()
		uc := NewPatchExpertProfile(f.experts, f.specs, f.publisher)

		view, err := uc.Execute(context.Background(), PatchCommand{AuthID: f.authID, ActorRole: profile.RoleAdmin, Bio: strPtr("new bio")})
		require.NoError(t, err)

		assert.Equal(t, "Dr. A", view.Name)
		assert.Equal(t, "0900", view.ExpertProfile.PhoneNumber)
		assert.Equal(t, "a@example.com", view.ExpertProfile.Email)
		assert.Equal(t, "new bio", view.ExpertProfile.Bio)
		assert.Empty(t, f.publisher.events)
	})

	t.Run("Admin đổi trạng thái xác minh và event được phát", func(t *testing.T) {
		f := newFixture()
		uc := NewPatchExpertProfile(f.experts, f.specs, f.publisher)

		view, err := uc.Execute(context.Background(), PatchCommand{
			AuthID: f.authID, ActorRole: profile.RoleAdmin, VerificationStatus: strPtr("REJECTED"),
		})
		require.NoError(t, err)
		assert.Equal(t, "REJECTED", view.ExpertProfile.VerificationStatus)
		require.Len(t, f.publisher.events, 1)
		assert.Equal(t, "expert.verification_status_changed", f.publisher.events[0].EventName())
	})

	t.Run("chuyên gia tự đổi trạng thái bị từ chối và không lưu gì", func(t *testing.T) {
		f := newFixture()
		uc := NewPatchExpertProfile(f.experts, f.specs, f.publisher)

		_, err := uc.Execute(context.Background(), PatchCommand{
			AuthID: f.authID, ActorRole: profile.RoleExpert, Bio: strPtr("x"), VerificationStatus: strPtr("VERIFIED"),
		})
		assert.ErrorIs(t, err, expert.ErrVerificationRequiresAdmin)
		assert.Nil(t, f.experts.saved)
	})

	t.Run("trạng thái không hợp lệ", func(t *testing.T) {
		f := newFixture()
		uc := NewPatchExpertProfile(f.experts, f.specs, f.publisher)

		_, err := uc.Execute(context.Background(), PatchCommand{
			AuthID: f.authID, ActorRole: profile.RoleAdmin, VerificationStatus: strPtr("APPROVED"),
		})
		assert.ErrorIs(t, err, expert.ErrInvalidVerificationStatus)
		assert.Nil(t, f.experts.saved)
	})
}
