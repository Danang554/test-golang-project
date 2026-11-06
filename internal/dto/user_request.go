package dto

// PaginationQuery handles query parameters for list endpoints
type PaginationQuery struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Search string `form:"search"`
}

// GetPage returns the current page with a default of 1
func (q *PaginationQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetLimit returns the page limit with a default of 10 and max of 100
func (q *PaginationQuery) GetLimit() int {
	if q.Limit <= 0 {
		return 10
	}
	if q.Limit > 100 {
		return 100
	}
	return q.Limit
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	IsAdmin  bool   `json:"is_admin"`
}

type UpdateUserRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=100"`
	Email    string  `json:"email" binding:"required,email,max=100"`
	Password string  `json:"password,omitempty" binding:"omitempty,min=6,max=100"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Email    string `json:"email,omitempty" binding:"omitempty,email,max=100"`
	Password string `json:"password,omitempty" binding:"omitempty,min=6,max=100"`
}
