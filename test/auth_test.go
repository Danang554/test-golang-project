package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"test-golang-project/config"
	"test-golang-project/internal/controller"
	"test-golang-project/internal/entity"
	"test-golang-project/internal/routes"
	"test-golang-project/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository implements repository.UserRepository for testing
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

// MockRevokedTokenRepository implements repository.RevokedTokenRepository for testing
type MockRevokedTokenRepository struct {
	mock.Mock
}

func (m *MockRevokedTokenRepository) Revoke(ctx context.Context, tokenHash string, userID uint, expiresAt time.Time) error {
	args := m.Called(ctx, tokenHash, userID, expiresAt)
	return args.Error(0)
}

func (m *MockRevokedTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
}

// setupAuthTestRouterWithRevocation initializes the test environment and exposes the revocation mock
func setupAuthTestRouterWithRevocation() (*gin.Engine, *MockUserRepository, *MockRevokedTokenRepository, *config.Config) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	mockRevokedTokenRepo := new(MockRevokedTokenRepository)
	mockRevokedTokenRepo.On("IsRevoked", mock.Anything, mock.Anything).Return(false, nil).Maybe()
	mockRevokedTokenRepo.On("Revoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	cfg := &config.Config{
		AppName: "test-auth-app",
		AppPort: "8080",
		AppEnv:  "testing",
		JWT: config.JWTConfig{
			Secret:          "test-secret-key-32-bytes-long!!",
			ExpirationHours: 24,
		},
	}

	authService := service.NewAuthService(mockRepo, mockRevokedTokenRepo, cfg)
	userService := service.NewUserService(mockRepo)

	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)

	router := routes.SetupRouter(routes.Config{
		AppConfig:        cfg,
		DB:               nil,
		AuthController:   authController,
		UserController:   userController,
		RevokedTokenRepo: mockRevokedTokenRepo,
	})

	return router, mockRepo, mockRevokedTokenRepo, cfg
}

// setupAuthTestRouter initializes the test environment with default mocks (preserves existing signature)
func setupAuthTestRouter() (*gin.Engine, *MockUserRepository, *config.Config) {
	router, mockRepo, _, cfg := setupAuthTestRouterWithRevocation()
	return router, mockRepo, cfg
}

// helper to create signed JWT for testing
func createTestToken(claims jwt.MapClaims, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte(secret))
	return signedToken
}

// =============================================================================
// ENDPOINT 1: POST /api/auth/register
// =============================================================================
func TestAuth_Register(t *testing.T) {
	// -------------------------------------------------------------------------
	// 1 KASUS SUKSES
	// -------------------------------------------------------------------------
	t.Run("Success - Register user with valid payload", func(t *testing.T) {
		router, mockRepo, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "John Doe",
			"email":    "john@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		// Mock: Email belum terdaftar, Create sukses
		mockRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Name == "John Doe" && u.Email == "john@example.com" && !u.IsAdmin
		})).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"email":"john@example.com"`)
		assert.NotContains(t, w.Body.String(), "password123") // Password tidak boleh bocor
		mockRepo.AssertExpectations(t)
	})

	// -------------------------------------------------------------------------
	// KASUS-KASUS GAGAL (SEBANYAK MUNGKIN)
	// -------------------------------------------------------------------------
	t.Run("Failure - Email already exists in database (409 Conflict)", func(t *testing.T) {
		router, mockRepo, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "Existing User",
			"email":    "existing@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		existing := &entity.User{ID: 1, Email: "existing@example.com"}
		mockRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existing, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Empty JSON body (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})

	t.Run("Failure - Malformed invalid JSON syntax (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{"name": "broken json`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})

	t.Run("Failure - Missing 'name' field (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"email":    "valid@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Name")
	})

	t.Run("Failure - Name too short, less than 2 characters (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "A", // Min is 2
			"email":    "valid@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Name too long, exceeds 100 characters (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     strings.Repeat("A", 101),
			"email":    "valid@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Missing 'email' field (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "Valid Name",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Invalid email format without at sign (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "Valid Name",
			"email":    "not-an-email-address",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Email too long, exceeds 100 characters (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "Valid Name",
			"email":    strings.Repeat("a", 95) + "@test.com", // > 100
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Missing 'password' field (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":  "Valid Name",
			"email": "valid@example.com",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Password too short, less than 6 characters (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "Valid Name",
			"email":    "valid@example.com",
			"password": "12345", // Min is 6
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Password too long, exceeds 100 characters (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "Valid Name",
			"email":    "valid@example.com",
			"password": strings.Repeat("p", 101), // Max is 100
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Database error on Create (500 Internal Server Error)", func(t *testing.T) {
		router, mockRepo, _ := setupAuthTestRouter()

		payload := map[string]string{
			"name":     "John Doe",
			"email":    "john@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		mockRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(errors.New("db connection failure")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		mockRepo.AssertExpectations(t)
	})
}

// =============================================================================
// ENDPOINT 2: POST /api/auth/login
// =============================================================================
func TestAuth_Login(t *testing.T) {
	// -------------------------------------------------------------------------
	// 1 KASUS SUKSES
	// -------------------------------------------------------------------------
	t.Run("Success - Login with valid credentials", func(t *testing.T) {
		router, mockRepo, _ := setupAuthTestRouter()

		rawPassword := "secretpassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
		user := &entity.User{
			ID:       1,
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: string(hashedPassword),
			IsAdmin:  false,
		}

		mockRepo.On("FindByEmail", mock.Anything, "alice@example.com").Return(user, nil).Once()

		payload := map[string]string{
			"email":    "alice@example.com",
			"password": rawPassword,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"token":`)
		assert.Contains(t, w.Body.String(), `"email":"alice@example.com"`)
		mockRepo.AssertExpectations(t)
	})

	// -------------------------------------------------------------------------
	// KASUS-KASUS GAGAL (SEBANYAK MUNGKIN)
	// -------------------------------------------------------------------------
	t.Run("Failure - Email not registered in system (401 Unauthorized)", func(t *testing.T) {
		router, mockRepo, _ := setupAuthTestRouter()

		mockRepo.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, nil).Once()

		payload := map[string]string{
			"email":    "notfound@example.com",
			"password": "anyPassword123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Incorrect password (401 Unauthorized)", func(t *testing.T) {
		router, mockRepo, _ := setupAuthTestRouter()

		correctPassword := "correctPassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
		user := &entity.User{
			ID:       1,
			Email:    "alice@example.com",
			Password: string(hashedPassword),
		}

		mockRepo.On("FindByEmail", mock.Anything, "alice@example.com").Return(user, nil).Once()

		payload := map[string]string{
			"email":    "alice@example.com",
			"password": "wrongPassword",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Empty JSON body (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"Email":"Email is required"`)
		assert.Contains(t, w.Body.String(), `"Password":"Password is required"`)
		assert.NotContains(t, w.Body.String(), "Key: 'LoginRequest")
	})

	t.Run("Failure - Malformed JSON (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email": "broken`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request payload format")
	})

	t.Run("Failure - Missing 'email' field (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{"password": "password123"}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"Email":"Email is required"`)
	})

	t.Run("Failure - Invalid email format (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{
			"email":    "notAnEmail",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"Email":"Email must be a valid email address"`)
	})

	t.Run("Failure - Missing 'password' field (400 Bad Request)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		payload := map[string]string{"email": "valid@example.com"}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"Password":"Password is required"`)
	})

	t.Run("Failure - Database error on FindByEmail (500 Internal Server Error)", func(t *testing.T) {
		router, mockRepo, _ := setupAuthTestRouter()

		mockRepo.On("FindByEmail", mock.Anything, "alice@example.com").Return(nil, errors.New("db error")).Once()

		payload := map[string]string{
			"email":    "alice@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		mockRepo.AssertExpectations(t)
	})
}

// =============================================================================
// ENDPOINT 3: POST /api/auth/logout
// =============================================================================
func TestAuth_Logout(t *testing.T) {
	// -------------------------------------------------------------------------
	// 1 KASUS SUKSES
	// -------------------------------------------------------------------------
	t.Run("Success - Logout with valid JWT token", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		user := &entity.User{ID: 10, Email: "user@example.com"}
		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()

		validToken := createTestToken(jwt.MapClaims{
			"sub":      10,
			"email":    "user@example.com",
			"is_admin": false,
			"exp":      time.Now().Add(time.Hour).Unix(),
			"iat":      time.Now().Unix(),
		}, cfg.JWT.Secret)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Logout successful")
		mockRepo.AssertExpectations(t)
	})

	// -------------------------------------------------------------------------
	// KASUS-KASUS GAGAL (SEBANYAK MUNGKIN)
	// -------------------------------------------------------------------------
	t.Run("Failure - Missing Authorization header (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is required")
	})

	t.Run("Failure - Invalid header format, missing Bearer prefix (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Basic some-basic-auth-value")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization format")
	})

	t.Run("Failure - Empty token string after Bearer (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Failure - Completely random/malformed token string (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer not-a-valid-jwt-token-string")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	t.Run("Failure - Expired token (401 Unauthorized)", func(t *testing.T) {
		router, _, cfg := setupAuthTestRouter()

		// exp 2 jam yang lalu
		expiredToken := createTestToken(jwt.MapClaims{
			"sub": 10,
			"exp": time.Now().Add(-2 * time.Hour).Unix(),
			"iat": time.Now().Add(-3 * time.Hour).Unix(),
		}, cfg.JWT.Secret)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	t.Run("Failure - Token signed with wrong secret key (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		wrongSecretToken := createTestToken(jwt.MapClaims{
			"sub": 10,
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}, "different-hacker-secret-key!!")

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+wrongSecretToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	t.Run("Failure - Valid token signature but missing 'sub' claim (401 Unauthorized)", func(t *testing.T) {
		router, _, cfg := setupAuthTestRouter()

		tokenWithoutSub := createTestToken(jwt.MapClaims{
			"email": "user@example.com",
			"exp":   time.Now().Add(time.Hour).Unix(),
		}, cfg.JWT.Secret)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+tokenWithoutSub)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token does not contain user identification")
	})

	t.Run("Failure - Valid token but 'sub' claim has invalid data type (401 Unauthorized)", func(t *testing.T) {
		router, _, cfg := setupAuthTestRouter()

		tokenWithInvalidSubType := createTestToken(jwt.MapClaims{
			"sub": "not-a-number-boolean-or-uint",
			"exp": time.Now().Add(time.Hour).Unix(),
		}, cfg.JWT.Secret)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+tokenWithInvalidSubType)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid user identification format")
	})

	t.Run("Failure - User ID from token does not exist in DB (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, nil).Once()

		token := createTestToken(jwt.MapClaims{
			"sub": 999,
			"exp": time.Now().Add(time.Hour).Unix(),
		}, cfg.JWT.Secret)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Database error when querying user during logout (500 Internal Server Error)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(nil, errors.New("database disconnected")).Once()

		token := createTestToken(jwt.MapClaims{
			"sub": 10,
			"exp": time.Now().Add(time.Hour).Unix(),
		}, cfg.JWT.Secret)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Already revoked token is rejected (401 Unauthorized)", func(t *testing.T) {
		router, _, mockRevokedTokenRepo, cfg := setupAuthTestRouterWithRevocation()

		token := createTestToken(jwt.MapClaims{
			"sub": 10,
			"exp": time.Now().Add(time.Hour).Unix(),
		}, cfg.JWT.Secret)

		// Mock: token is already marked as revoked in blacklist
		mockRevokedTokenRepo.ExpectedCalls = nil // reset defaults
		mockRevokedTokenRepo.On("IsRevoked", mock.Anything, mock.Anything).Return(true, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "Unauthorized. Please login first")
	})

	t.Run("Failure - Database error when revoking token during logout (500 Internal Server Error)", func(t *testing.T) {
		router, mockRepo, mockRevokedTokenRepo, cfg := setupAuthTestRouterWithRevocation()

		user := &entity.User{ID: 10, Email: "user@example.com"}
		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()

		token := createTestToken(jwt.MapClaims{
			"sub": 10,
			"exp": time.Now().Add(time.Hour).Unix(),
		}, cfg.JWT.Secret)

		// Mock: IsRevoked false, but Revoke fails with DB error
		mockRevokedTokenRepo.ExpectedCalls = nil
		mockRevokedTokenRepo.On("IsRevoked", mock.Anything, mock.Anything).Return(false, nil).Once()
		mockRevokedTokenRepo.On("Revoke", mock.Anything, mock.Anything, uint(10), mock.Anything).Return(errors.New("db insert failure")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})
}
