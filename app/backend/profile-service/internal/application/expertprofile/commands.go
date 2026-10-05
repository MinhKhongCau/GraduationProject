// File: internal/application/expertprofile/commands.go
package expertprofile

import (
	"profile-service/internal/domain/profile"

	"github.com/google/uuid"
)

// ReplaceCommand thay thế toàn bộ (PUT) hồ sơ chuyên gia. Không mang VerificationStatus:
// PUT luôn giữ nguyên trạng thái xác minh hiện tại.
type ReplaceCommand struct {
	AuthID               uuid.UUID
	Name                 string
	PhoneNumber          string
	Email                string
	AvatarURL            string
	IntroductionVideoURL string
	Bio                  string
	// SpecializationIDs: nil = giữ nguyên, rỗng = xoá hết.
	SpecializationIDs []string
}

// PatchCommand cập nhật một phần (PATCH) hồ sơ chuyên gia; field nil = không thay đổi.
type PatchCommand struct {
	AuthID               uuid.UUID
	ActorRole            profile.Role
	Name                 *string
	PhoneNumber          *string
	Email                *string
	AvatarURL            *string
	IntroductionVideoURL *string
	Bio                  *string
	VerificationStatus   *string
	// SpecializationIDs: nil = giữ nguyên, rỗng = xoá hết.
	SpecializationIDs []string
}
