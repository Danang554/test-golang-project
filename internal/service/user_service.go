package service

import (
	"context"
	"log/slog"
	"strings"

	"test-golang-project/internal/dto"
	"test-golang-project/internal/entity"
	"test-golang-project/internal/repository"
	"test-golang-project/pkg/apperror"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound        = apperror.NotFound("user not found")
	ErrDeletedUserNotFound = apperror.NotFound("deleted user not found")
)

type UserService interface {
	GetProfile(ctx context.Context, userID uint) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	GetUsers(ctx context.Context, query *dto.PaginationQuery) ([]*dto.UserListResponse, int64, error)
	GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	GetDeletedUsers(ctx context.Context, query *dto.PaginationQuery) ([]*dto.UserResponse, int64, error)
	RestoreUser(ctx context.Context, id uint) (*dto.UserResponse, error)
	ForceDeleteUser(ctx context.Context, id uint) error
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

func (s *userService) UpdateProfile(ctx context.Context, userID uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find user")
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	hasChanges := false

	// Update name if provided and not empty
	if strings.TrimSpace(req.Name) != "" {
		user.Name = strings.TrimSpace(req.Name)
		hasChanges = true
	}

	// Update email if provided, not empty, and different
	if strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Email) != user.Email {
		trimmedEmail := strings.TrimSpace(req.Email)
		existing, err := s.userRepo.FindByEmail(ctx, trimmedEmail)
		if err != nil {
			return nil, apperror.Internal(err, "Failed to check email")
		}
		if existing != nil && existing.ID != user.ID {
			return nil, apperror.Conflict("email is already registered")
		}
		user.Email = trimmedEmail
		hasChanges = true
	}

	// Update password if provided and not empty
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperror.Internal(err, "Failed to hash password")
		}
		user.Password = string(hashed)
		hasChanges = true
	}

	if hasChanges {
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, apperror.Internal(err, "Failed to update profile")
		}
		slog.Info("User updated their profile", "user_id", userID)
	}

	return dto.ToUserResponse(user), nil
}

func (s *userService) GetUsers(ctx context.Context, query *dto.PaginationQuery) ([]*dto.UserListResponse, int64, error) {
	page := query.GetPage()
	limit := query.GetLimit()

	users, total, err := s.userRepo.FindAll(ctx, query.Search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "Failed to retrieve users")
	}

	return dto.ToUserListResponse(users), total, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to retrieve user")
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return dto.ToUserResponse(user), nil
}

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to check email")
	}
	if existingUser != nil {
		return nil, apperror.Conflict("email is already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to hash password")
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsAdmin:  req.IsAdmin,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Internal(err, "Failed to create user")
	}

	slog.Info("Admin created new user", "user_id", user.ID, "email", user.Email, "is_admin", user.IsAdmin)
	return dto.ToUserResponse(user), nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find user")
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// If email is changed, check if new email is already taken by another user
	if req.Email != user.Email {
		existing, err := s.userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, apperror.Internal(err, "Failed to check email")
		}
		if existing != nil && existing.ID != user.ID {
			return nil, apperror.Conflict("email is already registered")
		}
		user.Email = req.Email
	}

	user.Name = req.Name

	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}

	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperror.Internal(err, "Failed to hash password")
		}
		user.Password = string(hashed)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(err, "Failed to update user")
	}

	slog.Info("User updated", "user_id", user.ID, "email", user.Email)
	return dto.ToUserResponse(user), nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return apperror.Internal(err, "Failed to find user")
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return apperror.Internal(err, "Failed to delete user")
	}

	slog.Info("User soft deleted", "user_id", id)
	return nil
}

func (s *userService) GetDeletedUsers(ctx context.Context, query *dto.PaginationQuery) ([]*dto.UserResponse, int64, error) {
	page := query.GetPage()
	limit := query.GetLimit()

	users, total, err := s.userRepo.FindAllDeleted(ctx, query.Search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "Failed to retrieve deleted users")
	}

	res := make([]*dto.UserResponse, 0, len(users))
	for _, u := range users {
		res = append(res, dto.ToUserResponse(u))
	}

	return res, total, nil
}

func (s *userService) RestoreUser(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindDeletedByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find deleted user")
	}
	if user == nil {
		return nil, ErrDeletedUserNotFound
	}

	if err := s.userRepo.Restore(ctx, id); err != nil {
		return nil, apperror.Internal(err, "Failed to restore user")
	}

	// Fetch refreshed user after restore
	restored, err := s.userRepo.FindByID(ctx, id)
	if err != nil || restored == nil {
		user.DeletedAt.Valid = false
		return dto.ToUserResponse(user), nil
	}

	slog.Info("User restored", "user_id", id)
	return dto.ToUserResponse(restored), nil
}

func (s *userService) ForceDeleteUser(ctx context.Context, id uint) error {
	user, err := s.userRepo.FindDeletedByID(ctx, id)
	if err != nil {
		return apperror.Internal(err, "Failed to find deleted user")
	}
	if user == nil {
		return ErrDeletedUserNotFound
	}

	if err := s.userRepo.ForceDelete(ctx, id); err != nil {
		return apperror.Internal(err, "Failed to permanently delete user")
	}

	slog.Info("User permanently deleted", "user_id", id)
	return nil
}
