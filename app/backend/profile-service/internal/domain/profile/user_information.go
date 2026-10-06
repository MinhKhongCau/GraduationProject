// File: internal/domain/profile/user_information.go
package profile

import "time"

// Gender là giới tính trong thông tin người dùng; chuỗi rỗng = chưa khai báo.
type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

// UserInformation là thông tin cá nhân dùng chung cho mọi vai trò (value object),
// lưu trực tiếp trên bảng profiles.
type UserInformation struct {
	FullName    string
	DateOfBirth *time.Time
	Gender      Gender
	PhoneNumber string
	Country     string
}

// UserInformationPatch mô tả cập nhật một phần; field nil = giữ nguyên.
// DateOfBirth không thể dùng nil để xoá nên có thêm cờ ClearDateOfBirth.
type UserInformationPatch struct {
	FullName         *string
	DateOfBirth      *time.Time
	ClearDateOfBirth bool
	Gender           *Gender
	PhoneNumber      *string
	Country          *string
}

// Apply trả về bản UserInformation mới sau khi áp dụng patch.
func (u UserInformation) Apply(p UserInformationPatch) UserInformation {
	if p.FullName != nil {
		u.FullName = *p.FullName
	}
	if p.ClearDateOfBirth {
		u.DateOfBirth = nil
	} else if p.DateOfBirth != nil {
		u.DateOfBirth = p.DateOfBirth
	}
	if p.Gender != nil {
		u.Gender = *p.Gender
	}
	if p.PhoneNumber != nil {
		u.PhoneNumber = *p.PhoneNumber
	}
	if p.Country != nil {
		u.Country = *p.Country
	}
	return u
}
