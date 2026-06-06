package schemas

type CreateSpecializationRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type SpecializationResponse struct {
	SpecID      string `json:"spec_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	IsActive    bool   `json:"is_active"`
}
