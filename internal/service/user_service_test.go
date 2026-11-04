package service_test

import (
	"context"
	"testing"

	"test-golang-project/config"
	"test-golang-project/internal/entity"
	"test-golang-project/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestUserService_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpirationHours: 24},
	}

	userService := service.NewUserService(mockRepo, cfg)

	req := &entity.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

	res, err := userService.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Email, res.Email)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", ExpirationHours: 24},
	}

	userService := service.NewUserService(mockRepo, cfg)

	req := &entity.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	existingUser := &entity.User{
		ID:    1,
		Name:  "Existing User",
		Email: req.Email,
	}
	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	res, err := userService.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, service.ErrEmailAlreadyExists)
	mockRepo.AssertExpectations(t)
}
