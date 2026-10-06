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
