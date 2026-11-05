package dto

import (
	"time"

	"test-golang-project/internal/entity"
)

type UserResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	IsAdmin   bool       `json:"is_admin"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	// DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UserListResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// ToUserResponse maps entity.User to UserResponse (Resource Mapper)
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

// ToUserListResponse maps entity.User to compact UserListResponse
func ToUserListResponse(user *entity.User) *UserListResponse {
	if user == nil {
		return nil
	}

	return &UserListResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	}
}
