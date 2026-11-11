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

func strPtr(s string) *string {
	return &s
}

// MockTodoRepository implements repository.TodoRepository for testing
type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) Create(ctx context.Context, todo *entity.Todo) error {
	args := m.Called(ctx, todo)
	return args.Error(0)
}

func (m *MockTodoRepository) FindByID(ctx context.Context, id uint) (*entity.Todo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Todo), args.Error(1)
}

func (m *MockTodoRepository) FindAllByUserID(ctx context.Context, userID uint, search string, categoryID *uint, isCompleted *bool, page int, limit int) ([]*entity.Todo, int64, error) {
	args := m.Called(ctx, userID, search, categoryID, isCompleted, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Todo), args.Get(1).(int64), args.Error(2)
}

func (m *MockTodoRepository) Update(ctx context.Context, todo *entity.Todo) error {
	args := m.Called(ctx, todo)
	return args.Error(0)
}

func (m *MockTodoRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTodoRepository) FindAllDeletedByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Todo, int64, error) {
	args := m.Called(ctx, userID, search, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Todo), args.Get(1).(int64), args.Error(2)
}

func (m *MockTodoRepository) FindDeletedByID(ctx context.Context, id uint) (*entity.Todo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Todo), args.Error(1)
}

func (m *MockTodoRepository) Restore(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTodoRepository) ForceDelete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTodoTestRouter() (*gin.Engine, *MockTodoRepository, *MockCategoryRepository, *config.Config) {
	gin.SetMode(gin.TestMode)

	mockUserRepo := new(MockUserRepository)
	mockRevokedTokenRepo := new(MockRevokedTokenRepository)
	mockRevokedTokenRepo.On("IsRevoked", mock.Anything, mock.Anything).Return(false, nil).Maybe()
	mockRevokedTokenRepo.On("Revoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	mockCategoryRepo := new(MockCategoryRepository)
	mockTodoRepo := new(MockTodoRepository)

	cfg := &config.Config{
		AppName: "test-todo-app",
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
	todoService := service.NewTodoService(mockTodoRepo, mockCategoryRepo)

	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)
	categoryController := controller.NewCategoryController(categoryService)
	todoController := controller.NewTodoController(todoService)

	router := routes.SetupRouter(routes.Config{
		AppConfig:          cfg,
		DB:                 nil,
		AuthController:     authController,
		UserController:     userController,
		CategoryController: categoryController,
		TodoController:     todoController,
		RevokedTokenRepo:   mockRevokedTokenRepo,
	})

	return router, mockTodoRepo, mockCategoryRepo, cfg
}

// =============================================================================
// TODO MANAGEMENT ENDPOINTS TESTS
// =============================================================================

func TestTodo_Authorization(t *testing.T) {
	t.Run("Failure - Unauthenticated request to /api/todo (401 Unauthorized)", func(t *testing.T) {
		router, _, _, _ := setupTodoTestRouter()

		req := httptest.NewRequest(http.MethodGet, "/api/todo", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "Unauthorized. Please login first")
	})
}

func TestTodo_Create(t *testing.T) {
	t.Run("Success - User creates todo without category", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		mockTodoRepo.On("Create", mock.Anything, mock.MatchedBy(func(todo *entity.Todo) bool {
			return todo.UserID == 10 && todo.Title == "Buy Groceries" && todo.CategoryID == nil && !todo.IsCompleted
		})).Return(nil).Once()

		payload := map[string]any{
			"title":       "Buy Groceries",
			"description": "Milk, Bread, Butter",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/todo", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Todo created successfully")
		assert.Contains(t, w.Body.String(), `"title":"Buy Groceries"`)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Success - User creates todo with valid category", func(t *testing.T) {
		router, mockTodoRepo, mockCategoryRepo, cfg := setupTodoTestRouter()

		catID := uint(5)
		category := &entity.Category{ID: catID, UserID: 10, Name: "Shopping"}
		mockCategoryRepo.On("FindByID", mock.Anything, catID).Return(category, nil).Once()

		createdTodo := &entity.Todo{
			ID:          1,
			UserID:      10,
			CategoryID:  &catID,
			Category:    category,
			Title:       "Buy Headphones",
			Description: strPtr("Noise cancelling"),
			IsCompleted: false,
		}
		mockTodoRepo.On("Create", mock.Anything, mock.MatchedBy(func(todo *entity.Todo) bool {
			return todo.UserID == 10 && todo.Title == "Buy Headphones" && todo.CategoryID != nil && *todo.CategoryID == 5
		})).Return(nil).Once()
		mockTodoRepo.On("FindByID", mock.Anything, mock.Anything).Return(createdTodo, nil).Once()

		payload := map[string]any{
			"title":       "Buy Headphones",
			"description": "Noise cancelling",
			"category_id": catID,
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/todo", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"title":"Buy Headphones"`)
		assert.Contains(t, w.Body.String(), `"name":"Shopping"`)
		mockCategoryRepo.AssertExpectations(t)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Failure - Category belongs to another user (400 Bad Request)", func(t *testing.T) {
		router, _, mockCategoryRepo, cfg := setupTodoTestRouter()

		catID := uint(99)
		// Belongs to user 20, but current user is 10
		otherCategory := &entity.Category{ID: catID, UserID: 20, Name: "Secret"}
		mockCategoryRepo.On("FindByID", mock.Anything, catID).Return(otherCategory, nil).Once()

		payload := map[string]any{
			"title":       "Illegal Todo",
			"category_id": catID,
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/todo", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "Category not found or does not belong to you")
		mockCategoryRepo.AssertExpectations(t)
	})

	t.Run("Failure - Missing required 'title' field (400 Bad Request)", func(t *testing.T) {
		router, _, _, cfg := setupTodoTestRouter()

		payload := map[string]any{
			"description": "No title here",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/todo", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})
}

func TestTodo_GetTodos(t *testing.T) {
	t.Run("Success - List todos with filter and pagination", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		catID := uint(5)
		isComp := false
		todos := []*entity.Todo{
			{ID: 1, UserID: 10, Title: "Learn Go", CategoryID: &catID, IsCompleted: false},
		}
		mockTodoRepo.On("FindAllByUserID", mock.Anything, uint(10), "learn", &catID, &isComp, 1, 10).
			Return(todos, int64(1), nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/todo?page=1&limit=10&search=learn&category_id=5&is_completed=false", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"total_items":1`)
		assert.Contains(t, w.Body.String(), `"Learn Go"`)
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodo_GetByID(t *testing.T) {
	t.Run("Success - Creator views their own todo (200 OK)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		todo := &entity.Todo{
			ID:          1,
			UserID:      10,
			Title:       "Personal Task",
			Description: strPtr("Details"),
			IsCompleted: false,
		}
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(todo, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/todo/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"Personal Task"`)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot view todo (404 Not Found)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		// Belongs to user 20
		todo := &entity.Todo{ID: 1, UserID: 20, Title: "User 20's Task"}
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(todo, nil).Once()

		// User 10 tries to view it
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/todo/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "todo not found")
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodo_Update(t *testing.T) {
	t.Run("Success - Creator updates todo title, clears description and category_id to null", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		existing := &entity.Todo{
			ID:          1,
			UserID:      10,
			Title:       "Old Title",
			IsCompleted: false,
		}
		updated := &entity.Todo{
			ID:          1,
			UserID:      10,
			Title:       "New Title",
			Description: nil,
			CategoryID:  nil,
			IsCompleted: true,
		}
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
		mockTodoRepo.On("Update", mock.Anything, mock.MatchedBy(func(td *entity.Todo) bool {
			return td.ID == 1 && td.Title == "New Title" && td.Description == nil && td.CategoryID == nil && td.IsCompleted == true
		})).Return(nil).Once()
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(updated, nil).Once()

		payload := map[string]any{
			"title":        "New Title",
			"description":  "",
			"category_id":  nil,
			"is_completed": true,
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Todo updated successfully")
		assert.Contains(t, w.Body.String(), `"title":"New Title"`)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Failure - Missing required 'title' field (400 Bad Request)", func(t *testing.T) {
		router, _, _, cfg := setupTodoTestRouter()

		payload := map[string]any{
			"description": "Updated without title",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "validation")
	})

	t.Run("Failure - Non-creator user cannot update todo (404 Not Found)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		// Belongs to user 20
		existing := &entity.Todo{ID: 1, UserID: 20, Title: "Someone else's todo"}
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()

		payload := map[string]any{"title": "Hacked Title"}
		body, _ := json.Marshal(payload)

		// User 10 tries to update it
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "todo not found")
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodo_ToggleComplete(t *testing.T) {
	t.Run("Success - Creator toggles uncompleted todo to completed via PATCH /api/todo/:id/complete", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		existing := &entity.Todo{
			ID:          1,
			UserID:      10,
			Title:       "Task to Complete",
			IsCompleted: false,
		}
		completed := &entity.Todo{
			ID:          1,
			UserID:      10,
			Title:       "Task to Complete",
			IsCompleted: true,
		}

		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
		mockTodoRepo.On("Update", mock.Anything, mock.MatchedBy(func(td *entity.Todo) bool {
			return td.ID == 1 && td.IsCompleted == true
		})).Return(nil).Once()
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(completed, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1/complete", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Todo marked as completed")
		assert.Contains(t, w.Body.String(), `"is_completed":true`)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Success - Creator toggles completed todo to uncompleted via PATCH /api/todo/:id/complete", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		existing := &entity.Todo{
			ID:          1,
			UserID:      10,
			Title:       "Task to Complete",
			IsCompleted: true,
		}
		uncompleted := &entity.Todo{
			ID:          1,
			UserID:      10,
			Title:       "Task to Complete",
			IsCompleted: false,
		}

		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
		mockTodoRepo.On("Update", mock.Anything, mock.MatchedBy(func(td *entity.Todo) bool {
			return td.ID == 1 && td.IsCompleted == false
		})).Return(nil).Once()
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(uncompleted, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1/complete", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Todo marked as uncompleted")
		assert.Contains(t, w.Body.String(), `"is_completed":false`)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot toggle todo (404 Not Found)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		// Belongs to user 20
		existing := &entity.Todo{ID: 1, UserID: 20, Title: "Other's Todo", IsCompleted: false}
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1/complete", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "todo not found")
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodo_Delete(t *testing.T) {
	t.Run("Success - Creator soft-deletes todo (200 OK)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		existing := &entity.Todo{ID: 1, UserID: 10, Title: "To Delete"}
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()
		mockTodoRepo.On("Delete", mock.Anything, uint(1)).Return(nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/todo/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Todo deleted successfully")
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot delete todo (404 Not Found)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		// Belongs to user 20
		existing := &entity.Todo{ID: 1, UserID: 20, Title: "Other's Todo"}
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(existing, nil).Once()

		// User 10 tries to delete it
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/todo/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "todo not found")
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodo_DeletedManagement(t *testing.T) {
	t.Run("Success - Creator lists deleted todos (trash)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		deletedTodos := []*entity.Todo{
			{
				ID:        1,
				UserID:    10,
				Title:     "Deleted Task",
				DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
			},
		}
		mockTodoRepo.On("FindAllDeletedByUserID", mock.Anything, uint(10), "", 1, 10).
			Return(deletedTodos, int64(1), nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/todo/deleted", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"Deleted Task"`)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Success - Creator restores soft-deleted todo via PATCH", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		deleted := &entity.Todo{
			ID:        1,
			UserID:    10,
			Title:     "To Restore",
			DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
		}
		restored := &entity.Todo{
			ID:        1,
			UserID:    10,
			Title:     "To Restore",
			DeletedAt: gorm.DeletedAt{Valid: false},
		}

		mockTodoRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()
		mockTodoRepo.On("Restore", mock.Anything, uint(1)).Return(nil).Once()
		mockTodoRepo.On("FindByID", mock.Anything, uint(1)).Return(restored, nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1/restore", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Todo restored successfully")
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot restore todo (404 Not Found)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		// Belongs to user 20
		deleted := &entity.Todo{
			ID:        1,
			UserID:    20,
			Title:     "Other's Deleted Todo",
			DeletedAt: gorm.DeletedAt{Valid: true},
		}
		mockTodoRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()

		// User 10 tries to restore it
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/todo/1/restore", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "deleted todo not found")
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Success - Creator permanently deletes todo from database", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		deleted := &entity.Todo{
			ID:        1,
			UserID:    10,
			DeletedAt: gorm.DeletedAt{Valid: true},
		}
		mockTodoRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()
		mockTodoRepo.On("ForceDelete", mock.Anything, uint(1)).Return(nil).Once()

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/todo/1/permanent", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Todo permanently deleted successfully")
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Failure - Non-creator user cannot permanently delete todo (404 Not Found)", func(t *testing.T) {
		router, mockTodoRepo, _, cfg := setupTodoTestRouter()

		deleted := &entity.Todo{
			ID:        1,
			UserID:    20,
			DeletedAt: gorm.DeletedAt{Valid: true},
		}
		mockTodoRepo.On("FindDeletedByID", mock.Anything, uint(1)).Return(deleted, nil).Once()

		// User 10 tries to force delete it
		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/todo/1/permanent", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "deleted todo not found")
		mockTodoRepo.AssertExpectations(t)
	})
}
