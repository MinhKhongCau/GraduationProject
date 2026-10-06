package schemas

// CreateSpecializationRequest tạo chuyên khoa mới; slug bỏ trống sẽ được sinh từ name.
type CreateSpecializationRequest struct {
	Code        string   `json:"code" binding:"required,max=50"`
	Name        string   `json:"name" binding:"required,max=255"`
	Slug        string   `json:"slug" binding:"omitempty,max=255"`
	Description string   `json:"description"`
	Symptoms    []string `json:"symptoms" binding:"omitempty,dive,required"`
	Location    string   `json:"location" binding:"omitempty,max=255"`
	ImageURL    string   `json:"image_url"`
}

// UpdateSpecializationRequest thay toàn bộ (PUT) thông tin chuyên khoa; slug bỏ trống sẽ sinh lại từ name.
// Trạng thái hoạt động đổi qua PATCH /specializations/{id}/status.
type UpdateSpecializationRequest struct {
	Code        string   `json:"code" binding:"required,max=50"`
	Name        string   `json:"name" binding:"required,max=255"`
	Slug        string   `json:"slug" binding:"omitempty,max=255"`
	Description string   `json:"description"`
	Symptoms    []string `json:"symptoms" binding:"omitempty,dive,required"`
	Location    string   `json:"location" binding:"omitempty,max=255"`
	ImageURL    string   `json:"image_url"`
}

// UpdateSpecializationStatusRequest bật/tắt chuyên khoa. Chuyên khoa tắt không hiện cho bệnh nhân
// và chuyên gia không đăng ký mới được, nhưng chuyên gia đã đăng ký vẫn giữ.
type UpdateSpecializationStatusRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}
