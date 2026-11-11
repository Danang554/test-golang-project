package dto

import (
	"time"

	"test-golang-project/internal/entity"
)

type TodoCategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type TodoResponse struct {
	ID          uint                  `json:"id"`
	Title       string                `json:"title"`
	Description *string               `json:"description"`
	IsCompleted bool                  `json:"is_completed"`
	Category    *TodoCategoryResponse `json:"category,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   *time.Time            `json:"deleted_at,omitempty"`
}

type TodoListResponse struct {
	ID          uint                  `json:"id"`
	Title       string                `json:"title"`
	Description *string               `json:"description"`
	IsCompleted bool                  `json:"is_completed"`
	CategoryID  *uint                 `json:"category_id"`
}

func ToTodoResponse(t *entity.Todo) *TodoResponse {
	if t == nil {
		return nil
	}

	var category *TodoCategoryResponse
	if t.Category != nil {
		category = &TodoCategoryResponse{
			ID:   t.Category.ID,
			Name: t.Category.Name,
		}
	}

	var deletedAt *time.Time
	if t.DeletedAt.Valid {
		deletedAt = &t.DeletedAt.Time
	}

	return &TodoResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		IsCompleted: t.IsCompleted,
		Category:    category,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}

func ToTodoListResponse(todos []*entity.Todo) []*TodoListResponse {
	res := make([]*TodoListResponse, 0, len(todos))
	for _, t := range todos {
		res = append(res, &TodoListResponse{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			IsCompleted: t.IsCompleted,
			CategoryID:  t.CategoryID,
		})
	}
	return res
}
