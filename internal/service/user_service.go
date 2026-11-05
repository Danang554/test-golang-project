package service

import (
	"context"

	"test-golang-project/internal/dto"
	"test-golang-project/internal/repository"
	"test-golang-project/pkg/apperror"
)

var (
	ErrUserNotFound = apperror.NotFound("user not found")
)

type UserService interface {
	GetProfile(ctx context.Context, userID uint) (*dto.UserResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to get profile")
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return dto.ToUserResponse(user), nil
}
