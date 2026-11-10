package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"test-golang-project/config"
	"test-golang-project/internal/controller"
	"test-golang-project/internal/entity"
	"test-golang-project/internal/routes"
	"test-golang-project/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockCategoryRepository implements repository.CategoryRepository for testing
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) FindByID(ctx context.Context, id uint) (*entity.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) FindAllByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Category, int64, error) {
	args := m.Called(ctx, userID, search, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Category), args.Get(1).(int64), args.Error(2)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) FindAllDeletedByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Category, int64, error) {
	args := m.Called(ctx, userID, search, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Category), args.Get(1).(int64), args.Error(2)
}

func (m *MockCategoryRepository) FindDeletedByID(ctx context.Context, id uint) (*entity.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) Restore(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) ForceDelete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupCategoryTestRouter() (*gin.Engine, *MockCategoryRepository, *config.Config) {
	gin.SetMode(gin.TestMode)

	mockUserRepo := new(MockUserRepository)
	mockRevokedTokenRepo := new(MockRevokedTokenRepository)
	mockRevokedTokenRepo.On("IsRevoked", mock.Anything, mock.Anything).Return(false, nil).Maybe()
	mockRevokedTokenRepo.On("Revoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	mockCategoryRepo := new(MockCategoryRepository)

	cfg := &config.Config{
		AppName: "test-category-app",
		AppPort: "8080",
		AppEnv:  "testing",
		JWT: config.JWTConfig{
			Secret:          "test-secret-key-32-bytes-long!!",
			ExpirationHours: 24,
		},
	}

	authService := service.NewAuthService(mockUserRepo, mockRevokedTokenRepo, cfg)
	userService := service.NewUserService(mockUserRepo)
	categoryService := service.NewCategoryService(mockCategoryRepo)

	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)
	categoryController := controller.NewCategoryController(categoryService)

	router := routes.SetupRouter(routes.Config{
		AppConfig:          cfg,
		DB:                 nil,
		AuthController:     authController,
		UserController:     userController,
		CategoryController: categoryController,
		RevokedTokenRepo:   mockRevokedTokenRepo,
	})

	return router, mockCategoryRepo, cfg
}

// =============================================================================
// CATEGORY MANAGEMENT ENDPOINTS TESTS
// =============================================================================

func TestCategory_Authorization(t *testing.T) {
	t.Run("Failure - Unauthenticated request to /api/category (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupCategoryTestRouter()

		req := httptest.NewRequest(http.MethodGet, "/api/category", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "Unauthorized. Please login first")
	})
}

func TestCategory_Create(t *testing.T) {
	t.Run("Success - User creates category", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
			return c.UserID == 10 && c.Name == "Technology" && c.Description == "Tech related items"
		})).Return(nil).Once()

		payload := map[string]string{
			"name":        "Technology",
			"description": "Tech related items",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/category", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Category created successfully")
		assert.Contains(t, w.Body.String(), `"name":"Technology"`)
		assert.NotContains(t, w.Body.String(), `"creator"`)
		assert.NotContains(t, w.Body.String(), `"user_id"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Missing required 'name' field (400 Bad Request)", func(t *testing.T) {
		router, _, cfg := setupCategoryTestRouter()

		payload := map[string]string{
			"description": "Tech without name",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/category", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})

	t.Run("Failure - Name too short, less than 2 characters (400 Bad Request)", func(t *testing.T) {
		router, _, cfg := setupCategoryTestRouter()

		payload := map[string]string{
			"name": "A",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/category", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})
}

func TestCategory_GetCategories(t *testing.T) {
	t.Run("Success - List categories belonging to the user with pagination and search", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		categories := []*entity.Category{
			{ID: 1, UserID: 10, Name: "Tech Gadgets", Description: "Electronic items"},
			{ID: 2, UserID: 10, Name: "Tech Books", Description: "Programming books"},
		}
		mockRepo.On("FindAllByUserID", mock.Anything, uint(10), "tech", 1, 10).
			Return(categories, int64(2), nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/category?page=1&limit=10&search=tech", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"total_items":2`)
		assert.Contains(t, w.Body.String(), `"Tech Gadgets"`)
		assert.NotContains(t, w.Body.String(), `"creator"`)
		assert.NotContains(t, w.Body.String(), `"user_id"`)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategory_GetByID(t *testing.T) {
	t.Run("Success - Creator views their own category (200 OK)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		category := &entity.Category{
			ID:          1,
			UserID:      10,
			Name:        "Personal Category",
			Description: "My personal stuff",
		}
		mockRepo.On("FindByID", mock.Anything, uint(1)).Return(category, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/category/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"Personal Category"`)
		assert.NotContains(t, w.Body.String(), `"creator"`)
		assert.NotContains(t, w.Body.String(), `"user_id"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Category not found (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/category/999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "category not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot view category (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		// Category created by user 20
		category := &entity.Category{
			ID:     1,
			UserID: 20,
			Name:   "User 20's Secret Category",
		}
		mockRepo.On("FindByID", mock.Anything, uint(1)).Return(category, nil).Once()

		// User 10 tries to view it -> Returns 404 Not Found (same as if it doesn't exist)
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/category/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "category not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestCategory_Update(t *testing.T) {
	t.Run("Success - Creator updates category via PATCH (200 OK)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		existing := &entity.Category{
			ID:          1,
			UserID:      10,
			Name:        "Old Category Name",
			Description: "Old Description",
		}
		mockRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
			return c.ID == 1 && c.Name == "New Category Name"
		})).Return(nil).Once()

		payload := map[string]string{
			"name": "New Category Name",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/category/1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"name":"New Category Name"`)
		assert.NotContains(t, w.Body.String(), `"creator"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot update category (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		// Belongs to user 20
		existing := &entity.Category{
			ID:     1,
			UserID: 20,
			Name:   "Someone Else's Category",
		}
		mockRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()

		payload := map[string]string{
			"name": "Hacked Name",
		}
		body, _ := json.Marshal(payload)

		// User 10 tries to update it -> Returns 404 Not Found
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/category/1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "category not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Category not found on update (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, nil).Once()

		payload := map[string]string{
			"name": "Updated Name",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/category/999", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "category not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestCategory_Delete(t *testing.T) {
	t.Run("Success - Creator soft-deletes category (200 OK)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		existing := &entity.Category{ID: 1, UserID: 10, Name: "Category to Delete"}
		mockRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
		mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/category/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Category deleted successfully")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot delete category (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		// Belongs to user 20
		existing := &entity.Category{ID: 1, UserID: 20, Name: "User 20's Category"}
		mockRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()

		// User 10 tries to delete it -> Returns 404 Not Found
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/category/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "category not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestCategory_DeletedManagement(t *testing.T) {
	t.Run("Success - Creator lists deleted categories (trash)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		deletedCategories := []*entity.Category{
			{
				ID:        1,
				UserID:    10,
				Name:      "Deleted Category",
				DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
			},
		}
		mockRepo.On("FindAllDeletedByUserID", mock.Anything, uint(10), "", 1, 10).
			Return(deletedCategories, int64(1), nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/category/deleted", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"Deleted Category"`)
		assert.NotContains(t, w.Body.String(), `"creator"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Creator restores soft-deleted category via PATCH", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		deleted := &entity.Category{
			ID:        1,
			UserID:    10,
			Name:      "To Restore",
			DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
		}
		restored := &entity.Category{
			ID:        1,
			UserID:    10,
			Name:      "To Restore",
			DeletedAt: gorm.DeletedAt{Valid: false},
		}

		mockRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()
		mockRepo.On("Restore", mock.Anything, uint(1)).Return(nil).Once()
		mockRepo.On("FindByID", mock.Anything, uint(1)).Return(restored, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/category/1/restore", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Category restored successfully")
		assert.NotContains(t, w.Body.String(), `"creator"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot restore category (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		// Belongs to user 20
		deleted := &entity.Category{
			ID:        1,
			UserID:    20,
			Name:      "User 20's Deleted Category",
			DeletedAt: gorm.DeletedAt{Valid: true},
		}
		mockRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()

		// User 10 tries to restore it -> Returns 404 Not Found
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/category/1/restore", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "deleted category not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Creator permanently deletes category from database", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		deleted := &entity.Category{
			ID:        1,
			UserID:    10,
			DeletedAt: gorm.DeletedAt{Valid: true},
		}
		mockRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()
		mockRepo.On("ForceDelete", mock.Anything, uint(1)).Return(nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/category/1/permanent", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Category permanently deleted successfully")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot permanently delete category (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupCategoryTestRouter()

		deleted := &entity.Category{
			ID:        1,
			UserID:    20,
			DeletedAt: gorm.DeletedAt{Valid: true},
		}
		mockRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()

		// User 10 tries to force delete it -> Returns 404 Not Found
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/category/1/permanent", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "deleted category not found")
		mockRepo.AssertExpectations(t)
	})
}
