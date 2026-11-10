package dto

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Description string `json:"description,omitempty" binding:"omitempty,max=255"`
}
