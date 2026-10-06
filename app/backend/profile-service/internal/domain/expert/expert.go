// File: internal/domain/expert/expert.go
package expert

import (
	"time"

	"profile-service/internal/domain/profile"
	"profile-service/internal/domain/specialization"

	"github.com/google/uuid"
)

// Details là phần thông tin hồ sơ chuyên gia có thể chỉnh sửa tự do (value object).
type Details struct {
	Email                string
	AvatarURL            string
	IntroductionVideoURL string
	Bio                  string
}

// Expert là aggregate root của hồ sơ chuyên gia (Profile vai trò EXPERT + ExpertProfile).
// Mọi thay đổi trạng thái đều đi qua method để giữ bất biến của domain.
type Expert struct {
	profileID          uuid.UUID
	authID             uuid.UUID
	slug               string
	userInformation    profile.UserInformation
	details            Details
	verificationStatus VerificationStatus
	specializations    []specialization.Specialization
	createdAt          time.Time
	updatedAt          time.Time

	specializationsChanged bool
	events                 []Event
}

// Snapshot là dạng dữ liệu phẳng để repository khôi phục/lưu aggregate mà không lộ field nội bộ.
type Snapshot struct {
	ProfileID          uuid.UUID
	AuthID             uuid.UUID
	Slug               string
	UserInformation    profile.UserInformation
	Details            Details
	VerificationStatus string
	Specializations    []specialization.Specialization
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Reconstitute khôi phục Expert từ dữ liệu đã lưu. Profile EXPERT chưa có bản ghi ExpertProfile
// (VerificationStatus rỗng) được coi là UNVERIFIED.
func Reconstitute(s Snapshot) *Expert {
	return &Expert{
		profileID:          s.ProfileID,
		authID:             s.AuthID,
		slug:               s.Slug,
		userInformation:    s.UserInformation,
		details:            s.Details,
		verificationStatus: verificationStatusFromStorage(s.VerificationStatus),
		specializations:    s.Specializations,
		createdAt:          s.CreatedAt,
		updatedAt:          s.UpdatedAt,
	}
}

func (e *Expert) Snapshot() Snapshot {
	return Snapshot{
		ProfileID:          e.profileID,
		AuthID:             e.authID,
		Slug:               e.slug,
		UserInformation:    e.userInformation,
		Details:            e.details,
		VerificationStatus: string(e.verificationStatus),
		Specializations:    e.specializations,
		CreatedAt:          e.createdAt,
		UpdatedAt:          e.updatedAt,
	}
}

func (e *Expert) UserInformation() profile.UserInformation         { return e.userInformation }
func (e *Expert) Details() Details                                 { return e.details }
func (e *Expert) VerificationStatus() VerificationStatus           { return e.verificationStatus }
func (e *Expert) Specializations() []specialization.Specialization { return e.specializations }

// SpecializationsChanged cho repository biết có cần đồng bộ lại bảng expert_specializations hay không.
func (e *Expert) SpecializationsChanged() bool { return e.specializationsChanged }

// ReviseUserInformation thay thông tin cá nhân (họ tên, ngày sinh, giới tính, SĐT, quốc gia).
func (e *Expert) ReviseUserInformation(info profile.UserInformation) {
	e.userInformation = info
}

// ReviseDetails thay toàn bộ thông tin hồ sơ. Không đụng tới trạng thái xác minh: trạng thái chỉ
// đổi được qua ChangeVerificationStatus.
func (e *Expert) ReviseDetails(d Details) {
	e.details = d
}

// ChangeVerificationStatus đổi trạng thái xác minh; chỉ Admin được phép.
func (e *Expert) ChangeVerificationStatus(status VerificationStatus, actor profile.Role, at time.Time) error {
	if !actor.IsAdmin() {
		return ErrVerificationRequiresAdmin
	}
	if _, err := ParseVerificationStatus(string(status)); err != nil {
		return err
	}
	if status == e.verificationStatus {
		return nil
	}

	e.events = append(e.events, VerificationStatusChanged{
		ProfileID: e.profileID,
		AuthID:    e.authID,
		From:      e.verificationStatus,
		To:        status,
		At:        at,
	})
	e.verificationStatus = status
	return nil
}

// AssignSpecializations thay toàn bộ danh sách chuyên khoa (danh sách rỗng = xoá hết).
func (e *Expert) AssignSpecializations(specs []specialization.Specialization) {
	e.specializations = specs
	e.specializationsChanged = true
}

// PullEvents trả về và xoá các domain event đang chờ phát.
func (e *Expert) PullEvents() []Event {
	events := e.events
	e.events = nil
	return events
}
