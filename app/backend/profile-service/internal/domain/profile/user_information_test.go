package profile

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserInformation_Apply(t *testing.T) {
	dob := time.Date(1995, 5, 15, 0, 0, 0, 0, time.UTC)
	base := UserInformation{FullName: "A", DateOfBirth: &dob, Gender: GenderMale, PhoneNumber: "0900", Country: "Vietnam"}

	t.Run("patch rỗng giữ nguyên", func(t *testing.T) {
		assert.Equal(t, base, base.Apply(UserInformationPatch{}))
	})

	t.Run("chỉ đổi field được gửi", func(t *testing.T) {
		name, country := "B", "Japan"
		got := base.Apply(UserInformationPatch{FullName: &name, Country: &country})
		assert.Equal(t, "B", got.FullName)
		assert.Equal(t, "Japan", got.Country)
		assert.Equal(t, "0900", got.PhoneNumber)
		assert.Equal(t, &dob, got.DateOfBirth)
	})

	t.Run("xoá ngày sinh", func(t *testing.T) {
		assert.Nil(t, base.Apply(UserInformationPatch{ClearDateOfBirth: true}).DateOfBirth)
	})
}
