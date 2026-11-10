package dto

import (
	"time"

	"test-golang-project/internal/entity"
)

// UserResponse represents full user data (detail, profile, create, update)
type UserResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	IsAdmin   bool       `json:"is_admin"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserListResponse represents compact user data for list endpoints (no timestamps)
type UserListResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// ToUserResponse maps a single entity.User to full UserResponse
func ToUserResponse(user *entity.User) *UserResponse {
	if user == nil {
		return nil
	}

	var deletedAt *time.Time
	if user.DeletedAt.Valid {
		deletedAt = &user.DeletedAt.Time
	}

	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// ToUserListResponse maps a slice of entity.User to compact UserListResponse slice
func ToUserListResponse(users []*entity.User) []*UserListResponse {
	list := make([]*UserListResponse, 0, len(users))
	for _, u := range users {
		if u == nil {
			continue
		}
		list = append(list, &UserListResponse{
			ID:      u.ID,
			Name:    u.Name,
			Email:   u.Email,
			IsAdmin: u.IsAdmin,
		})
	}
	return list
}
