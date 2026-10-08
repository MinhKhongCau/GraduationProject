// File: internal/infrastructure/http/handlers/user_information.go
package handlers

import (
	"errors"
	"net/http"

	"profile-service/internal/domain/profile"
	"profile-service/internal/infrastructure/http/response"
	"profile-service/internal/infrastructure/http/schemas"
	"profile-service/internal/infrastructure/persistence/models"

	"github.com/gin-gonic/gin"
)

var errInvalidDateOfBirth = errors.New("invalid date_of_birth, expected YYYY-MM-DD")

// parseUserInformation map request PUT sang value object của domain (field không gửi = rỗng).
func parseUserInformation(req schemas.UserInformationRequest) (profile.UserInformation, error) {
	dob, err := parseOptionalDate(req.DateOfBirth)
	if err != nil {
		return profile.UserInformation{}, errInvalidDateOfBirth
	}
	return profile.UserInformation{
		FullName:    req.FullName,
		DateOfBirth: dob,
		Gender:      profile.Gender(req.Gender),
		PhoneNumber: req.PhoneNumber,
		Country:     req.Country,
	}, nil
}

// parseUserInformationPatch map request PATCH sang patch của domain; req nil = không đổi gì.
func parseUserInformationPatch(req *schemas.PatchUserInformationRequest) (profile.UserInformationPatch, error) {
	if req == nil {
		return profile.UserInformationPatch{}, nil
	}
	patch := profile.UserInformationPatch{
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Country:     req.Country,
	}
	if req.Gender != nil {
		gender := profile.Gender(*req.Gender)
		patch.Gender = &gender
	}
	if req.DateOfBirth != nil {
		dob, err := parseOptionalDate(*req.DateOfBirth)
		if err != nil {
			return profile.UserInformationPatch{}, errInvalidDateOfBirth
		}
		patch.DateOfBirth = dob
		patch.ClearDateOfBirth = dob == nil
	}
	return patch, nil
}

func writeInvalidDateOfBirth(c *gin.Context, err error) {
	response.Error(c, http.StatusBadRequest, "Định dạng ngày sinh không hợp lệ (YYYY-MM-DD)", err.Error())
}

// sanitizePublicProfile ẩn thông tin cá nhân khi trả profile cho người khác xem.
// Chuyên gia giữ nguyên thông tin vì đây là hồ sơ công khai để bệnh nhân lựa chọn.
// Riêng Admin quản lý chuyên gia là thông tin nội bộ nên luôn bị ẩn.
func sanitizePublicProfile(p *models.Profile) {
	if p.Role == models.RoleExpert {
		if p.ExpertProfile != nil {
			p.ExpertProfile.ManagedByAdminID = nil
		}
		return
	}
	p.UserInformation = models.UserInformation{FullName: p.UserInformation.FullName}
	if p.PatientProfile != nil {
		p.PatientProfile.Email = ""
		p.PatientProfile.Address = ""
		p.PatientProfile.MedicalHistories = nil
	}
	if p.AdminProfile != nil {
		p.AdminProfile.Email = ""
	}
}
