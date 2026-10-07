package schemas

// CreatePatientRecordRequest là hồ sơ người thân mà bệnh nhân tạo để đặt lịch hộ.
// Hồ sơ SELF được sinh tự động từ /profiles/me nên không tạo qua API này.
type CreatePatientRecordRequest struct {
	FullName     string `json:"full_name" binding:"required,max=255"`
	DateOfBirth  string `json:"date_of_birth" binding:"required"` // Định dạng: YYYY-MM-DD
	Gender       string `json:"gender" binding:"required,oneof=MALE FEMALE OTHER"`
	PhoneNumber  string `json:"phone_number" binding:"required,max=20"`
	Email        string `json:"email" binding:"omitempty,email,max=255"`
	Address      string `json:"address" binding:"omitempty,max=500"`
	Relationship string `json:"relationship" binding:"required,oneof=PARENT CHILD SPOUSE SIBLING RELATIVE OTHER"`
}
