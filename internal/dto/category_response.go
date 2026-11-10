package dto

import (
	"time"

	"test-golang-project/internal/entity"
)

type CategoryResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CategoryListResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ToCategoryResponse(c *entity.Category) *CategoryResponse {
	if c == nil {
		return nil
	}

	var deletedAt *time.Time
	if c.DeletedAt.Valid {
		deletedAt = &c.DeletedAt.Time
	}

	return &CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}

func ToCategoryListResponse(categories []*entity.Category) []*CategoryListResponse {
	res := make([]*CategoryListResponse, 0, len(categories))
	for _, c := range categories {
		res = append(res, &CategoryListResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
		})
	}
	return res
}
