package dto

type TodoQuery struct {
	PaginationQuery
	CategoryID  *uint `form:"category_id"`
	IsCompleted *bool `form:"is_completed"`
}

type CreateTodoRequest struct {
	Title       string  `json:"title" binding:"required,min=2,max=255"`
	Description *string `json:"description" binding:"omitempty"`
	CategoryID  *uint   `json:"category_id" binding:"omitempty"`
}

type UpdateTodoRequest struct {
	Title       string  `json:"title" binding:"required,min=2,max=255"`
	Description *string `json:"description"`
	CategoryID  *uint   `json:"category_id"`
	IsCompleted *bool   `json:"is_completed,omitempty"`
}
