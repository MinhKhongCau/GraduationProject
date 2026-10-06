package expertprofile

import (
	"context"
	"testing"

	"profile-service/internal/domain/expert"
	"profile-service/internal/domain/specialization"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManageExpertSpecializations(t *testing.T) {
	t.Run("thêm chuyên khoa mới", func(t *testing.T) {
		f := newFixture()
		uc := NewManageExpertSpecializations(f.experts, f.specs, f.publisher)

		view, err := uc.Add(context.Background(), f.authID, f.spec.ID)
		require.NoError(t, err)
		assert.Len(t, view.ExpertProfile.Specializations, 2)
		assert.True(t, f.experts.saved.SpecializationsChanged())
	})

	t.Run("thêm chuyên khoa đã có là idempotent", func(t *testing.T) {
		f := newFixture()
		uc := NewManageExpertSpecializations(f.experts, f.specs, f.publisher)
		existing := f.experts.experts[f.authID].Specializations[0]
		f.specs.specs[existing.ID.String()] = existing

		view, err := uc.Add(context.Background(), f.authID, existing.ID)
		require.NoError(t, err)
		assert.Len(t, view.ExpertProfile.Specializations, 1)
		assert.False(t, f.experts.saved.SpecializationsChanged())
	})

	t.Run("không thêm được chuyên khoa không tồn tại hoặc ngừng hoạt động", func(t *testing.T) {
		f := newFixture()
		inactive := specialization.Specialization{ID: uuid.New(), IsActive: false}
		f.specs.specs[inactive.ID.String()] = inactive
		uc := NewManageExpertSpecializations(f.experts, f.specs, f.publisher)

		_, err := uc.Add(context.Background(), f.authID, uuid.New())
		assert.ErrorIs(t, err, specialization.ErrSpecializationNotFound)
		_, err = uc.Add(context.Background(), f.authID, inactive.ID)
		assert.ErrorIs(t, err, specialization.ErrSpecializationInactive)
		assert.Nil(t, f.experts.saved)
	})

	t.Run("bỏ chuyên khoa", func(t *testing.T) {
		f := newFixture()
		uc := NewManageExpertSpecializations(f.experts, f.specs, f.publisher)
		existing := f.experts.experts[f.authID].Specializations[0]

		view, err := uc.Remove(context.Background(), f.authID, existing.ID)
		require.NoError(t, err)
		assert.Empty(t, view.ExpertProfile.Specializations)

		_, err = uc.Remove(context.Background(), f.authID, f.spec.ID)
		assert.ErrorIs(t, err, expert.ErrSpecializationNotAssigned)
	})

	t.Run("không tìm thấy chuyên gia", func(t *testing.T) {
		f := newFixture()
		uc := NewManageExpertSpecializations(f.experts, f.specs, f.publisher)

		_, err := uc.Add(context.Background(), uuid.New(), f.spec.ID)
		assert.ErrorIs(t, err, expert.ErrExpertNotFound)
	})
}
