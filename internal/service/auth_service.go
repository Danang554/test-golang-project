package service

import (
	"context"
	"log/slog"
	"time"

	"test-golang-project/config"
	"test-golang-project/internal/dto"
	"test-golang-project/internal/entity"
	"test-golang-project/internal/repository"
	"test-golang-project/pkg/apperror"
	"test-golang-project/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = apperror.Conflict("email is already registered")
	ErrInvalidCredentials = apperror.Unauthorized("invalid email or password")
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Logout(ctx context.Context, tokenString string, userID uint, expiresAt time.Time) error
}

type authService struct {
	userRepo         repository.UserRepository
	revokedTokenRepo repository.RevokedTokenRepository
	cfg              *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	revokedTokenRepo repository.RevokedTokenRepository,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		revokedTokenRepo: revokedTokenRepo,
		cfg:              cfg,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	// Check if email is already taken
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to register user")
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to hash password")
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsAdmin:  false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Internal(err, "Failed to register user")
	}

	slog.Info("New user registered", "user_id", user.ID, "email", user.Email)
	return dto.ToUserResponse(user), nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to login")
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to generate token")
	}

	slog.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)
	return &dto.LoginResponse{
		Token: token,
		User:  *dto.ToUserResponse(user),
	}, nil
}

func (s *authService) Logout(ctx context.Context, tokenString string, userID uint, expiresAt time.Time) error {
	// Verify user exists in database
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.Internal(err, "Failed to logout")
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Blacklist the token if revokedTokenRepo is available
	if s.revokedTokenRepo != nil && tokenString != "" {
		tokenHash := utils.HashToken(tokenString)
		if expiresAt.IsZero() {
			expiresAt = time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.ExpirationHours))
		}
		if err := s.revokedTokenRepo.Revoke(ctx, tokenHash, userID, expiresAt); err != nil {
			return apperror.Internal(err, "Failed to revoke token")
		}
	}

	slog.Info("User logged out successfully", "user_id", userID, "email", user.Email)
	return nil
}

func (s *authService) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"email":    user.Email,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.ExpirationHours)).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}
