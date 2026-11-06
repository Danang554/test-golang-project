package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"test-golang-project/internal/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// helper to create admin JWT token
func createAdminTestToken(userID uint, secret string) string {
	return createTestToken(jwt.MapClaims{
		"sub":      userID,
		"is_admin": true,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}, secret)
}

// helper to create non-admin member JWT token
func createMemberTestToken(userID uint, secret string) string {
	return createTestToken(jwt.MapClaims{
		"sub":      userID,
		"is_admin": false,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}, secret)
}

// =============================================================================
// USER MANAGEMENT ENDPOINTS TESTS
// =============================================================================

func TestUser_Authorization(t *testing.T) {
	t.Run("Failure - Unauthenticated request to /api/user (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})

	t.Run("Failure - Non-admin user denied access to /api/user (403 Forbidden)", func(t *testing.T) {
		router, _, cfg := setupAuthTestRouter()

		memberToken := createMemberTestToken(2, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		req.Header.Set("Authorization", "Bearer "+memberToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "you do not have permission to access this resource")
	})

	t.Run("Success - Admin user allowed access to /api/user (200 OK)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		mockRepo.On("FindAll", mock.Anything, "", 1, 10).Return([]*entity.User{}, int64(0), nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"pagination":`)
		mockRepo.AssertExpectations(t)
	})
}

func TestUser_GetUsers_PaginationAndSearch(t *testing.T) {
	t.Run("Success - List active users with pagination and search", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		users := []*entity.User{
			{ID: 1, Name: "Admin User", Email: "admin@example.com", IsAdmin: true},
			{ID: 2, Name: "John Doe", Email: "john@example.com", IsAdmin: false},
		}
		mockRepo.On("FindAll", mock.Anything, "john", 2, 5).Return(users, int64(12), nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/user?page=2&limit=5&search=john", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"total_items":12`)
		assert.Contains(t, w.Body.String(), `"total_pages":3`)
		assert.Contains(t, w.Body.String(), `"page":2`)
		assert.Contains(t, w.Body.String(), `"limit":5`)
		assert.Contains(t, w.Body.String(), `"john@example.com"`)
		assert.NotContains(t, w.Body.String(), `"created_at"`)
		assert.NotContains(t, w.Body.String(), `"updated_at"`)
		assert.NotContains(t, w.Body.String(), `"deleted_at"`)
		mockRepo.AssertExpectations(t)
	})
}

func TestUser_CreateUser(t *testing.T) {
	t.Run("Success - Admin creates new user", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		payload := map[string]any{
			"name":     "New Person",
			"email":    "newperson@example.com",
			"password": "securepassword",
			"is_admin": false,
		}
		body, _ := json.Marshal(payload)

		mockRepo.On("FindByEmail", mock.Anything, "newperson@example.com").Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Name == "New Person" && u.Email == "newperson@example.com" && !u.IsAdmin
		})).Return(nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/user", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"email":"newperson@example.com"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Duplicate email conflict (409 Conflict)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		payload := map[string]any{
			"name":     "Duplicate",
			"email":    "existing@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		existing := &entity.User{ID: 5, Email: "existing@example.com"}
		mockRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existing, nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPost, "/api/user", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "email is already registered")
		mockRepo.AssertExpectations(t)
	})
}

func TestUser_UpdateUser(t *testing.T) {
	t.Run("Success - Update user details", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		existing := &entity.User{ID: 2, Name: "Old Name", Email: "old@example.com", IsAdmin: false}
		mockRepo.On("FindByID", mock.Anything, uint(2)).Return(existing, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Name == "Updated Name" && u.Email == "old@example.com"
		})).Return(nil).Once()

		payload := map[string]any{
			"name":  "Updated Name",
			"email": "old@example.com",
		}
		body, _ := json.Marshal(payload)

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/user/2", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"Updated Name"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Update user not found (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, nil).Once()

		payload := map[string]any{
			"name":  "Name",
			"email": "test@example.com",
		}
		body, _ := json.Marshal(payload)

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/user/999", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestUser_SoftDelete(t *testing.T) {
	t.Run("Success - Soft delete user", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		existing := &entity.User{ID: 3, Email: "user3@example.com"}
		mockRepo.On("FindByID", mock.Anything, uint(3)).Return(existing, nil).Once()
		mockRepo.On("Delete", mock.Anything, uint(3)).Return(nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/user/3", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User deleted successfully")
		mockRepo.AssertExpectations(t)
	})
}

func TestUser_DeletedManagement(t *testing.T) {
	t.Run("Success - List deleted users (trash)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		deletedTime := time.Now()
		deletedUsers := []*entity.User{
			{ID: 4, Name: "Deleted User", Email: "deleted@example.com", DeletedAt: gorm.DeletedAt{Time: deletedTime, Valid: true}},
		}
		mockRepo.On("FindAllDeleted", mock.Anything, "", 1, 10).Return(deletedUsers, int64(1), nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodGet, "/api/user/deleted", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"deleted@example.com"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Restore soft-deleted user", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		deletedUser := &entity.User{ID: 4, Email: "restore@example.com", DeletedAt: gorm.DeletedAt{Valid: true}}
		restoredUser := &entity.User{ID: 4, Email: "restore@example.com", DeletedAt: gorm.DeletedAt{Valid: false}}

		mockRepo.On("FindDeletedByID", mock.Anything, uint(4)).Return(deletedUser, nil).Once()
		mockRepo.On("Restore", mock.Anything, uint(4)).Return(nil).Once()
		mockRepo.On("FindByID", mock.Anything, uint(4)).Return(restoredUser, nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/user/4/restore", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "User restored successfully")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Permanently delete user from database", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		deletedUser := &entity.User{ID: 5, Email: "permanent@example.com", DeletedAt: gorm.DeletedAt{Valid: true}}
		mockRepo.On("FindDeletedByID", mock.Anything, uint(5)).Return(deletedUser, nil).Once()
		mockRepo.On("ForceDelete", mock.Anything, uint(5)).Return(nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/user/5/permanent", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "User permanently deleted successfully")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Permanently delete non-existent deleted user (404 Not Found)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		mockRepo.On("FindDeletedByID", mock.Anything, uint(888)).Return(nil, nil).Once()

		adminToken := createAdminTestToken(1, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodDelete, "/api/user/888/permanent", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "deleted user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestUser_UpdateProfile(t *testing.T) {
	t.Run("Success - Update name only", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		user := &entity.User{
			ID:    10,
			Name:  "Old Name",
			Email: "user@example.com",
		}

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == 10 && u.Name == "New Name" && u.Email == "user@example.com"
		})).Return(nil).Once()

		payload := map[string]string{
			"name": "New Name",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Profile updated successfully")
		assert.Contains(t, w.Body.String(), `"name":"New Name"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Update password only", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		rawCurrent := "oldpassword123"
		hashedCurrent, _ := bcrypt.GenerateFromPassword([]byte(rawCurrent), bcrypt.DefaultCost)
		user := &entity.User{
			ID:       10,
			Name:     "Existing Name",
			Email:    "user@example.com",
			Password: string(hashedCurrent),
		}

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == 10 && u.Password != string(hashedCurrent) && u.Name == "Existing Name"
		})).Return(nil).Once()

		payload := map[string]string{
			"password": "brandnewpassword",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), "Profile updated successfully")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Update both name and password", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		user := &entity.User{
			ID:       10,
			Name:     "Old Name",
			Email:    "user@example.com",
			Password: "oldhashedpassword",
		}

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == 10 && u.Name == "Updated Name" && u.Password != "oldhashedpassword"
		})).Return(nil).Once()

		payload := map[string]string{
			"name":     "Updated Name",
			"password": "newsecurepassword",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"name":"Updated Name"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - No changes provided (empty payload)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		user := &entity.User{
			ID:    10,
			Name:  "Current Name",
			Email: "user@example.com",
		}

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
		// Update should not be called

		payload := map[string]string{}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"name":"Current Name"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Payload fields prepared with empty strings does not trigger validation error", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		user := &entity.User{
			ID:    10,
			Name:  "Current Name",
			Email: "user@example.com",
		}

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
		// Update should not be called because all fields are empty strings

		payload := map[string]string{
			"name":     "",
			"email":    "",
			"password": "",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"name":"Current Name"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Update name while password and email are empty strings in payload", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		user := &entity.User{
			ID:       10,
			Name:     "Old Name",
			Email:    "user@example.com",
			Password: "existinghashedpassword",
		}

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(user, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == 10 && u.Name == "Brand New Name" && u.Password == "existinghashedpassword" && u.Email == "user@example.com"
		})).Return(nil).Once()

		payload := map[string]string{
			"name":     "Brand New Name",
			"email":    "",
			"password": "",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"name":"Brand New Name"`)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Validation error on short password (400 Bad Request)", func(t *testing.T) {
		router, _, cfg := setupAuthTestRouter()

		payload := map[string]string{
			"password": "123", // too short (< 6)
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})

	t.Run("Failure - Email conflict with another user (409 Conflict)", func(t *testing.T) {
		router, mockRepo, cfg := setupAuthTestRouter()

		currentUser := &entity.User{
			ID:    10,
			Name:  "Current User",
			Email: "current@example.com",
		}
		otherUser := &entity.User{
			ID:    20,
			Name:  "Other User",
			Email: "taken@example.com",
		}

		mockRepo.On("FindByID", mock.Anything, uint(10)).Return(currentUser, nil).Once()
		mockRepo.On("FindByEmail", mock.Anything, "taken@example.com").Return(otherUser, nil).Once()

		payload := map[string]string{
			"email": "taken@example.com",
		}
		body, _ := json.Marshal(payload)

		token := createMemberTestToken(10, cfg.JWT.Secret)
		req := httptest.NewRequest(http.MethodPatch, "/api/profile", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), "email is already registered")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Unauthenticated request to update profile (401 Unauthorized)", func(t *testing.T) {
		router, _, _ := setupAuthTestRouter()

		req := httptest.NewRequest(http.MethodPatch, "/api/profile", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})
}
