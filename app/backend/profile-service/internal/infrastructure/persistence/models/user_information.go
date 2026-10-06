// File: internal/infrastructure/persistence/models/user_information.go
package models

import "profile-service/internal/domain/profile"

// NewUserInformation map value object của domain sang các cột trên bảng profiles.
func NewUserInformation(u profile.UserInformation) UserInformation {
	return UserInformation{
		FullName:    u.FullName,
		DateOfBirth: u.DateOfBirth,
		Gender:      string(u.Gender),
		PhoneNumber: u.PhoneNumber,
		Country:     u.Country,
	}
}

func (u UserInformation) ToDomain() profile.UserInformation {
	return profile.UserInformation{
		FullName:    u.FullName,
		DateOfBirth: u.DateOfBirth,
		Gender:      profile.Gender(u.Gender),
		PhoneNumber: u.PhoneNumber,
		Country:     u.Country,
	}
}

// Columns trả về map cột -> giá trị để UPDATE, ghi cả giá trị rỗng/NULL (khác Updates(struct) của GORM).
func (u UserInformation) Columns() map[string]any {
	return map[string]any{
		"full_name":     u.FullName,
		"date_of_birth": u.DateOfBirth,
		"gender":        u.Gender,
		"phone_number":  u.PhoneNumber,
		"country":       u.Country,
	}
}
