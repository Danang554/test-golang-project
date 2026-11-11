package service

import (
	"context"
	"log/slog"
	"strings"

	"test-golang-project/internal/dto"
	"test-golang-project/internal/entity"
	"test-golang-project/internal/repository"
	"test-golang-project/pkg/apperror"
)

var (
	ErrTodoNotFound        = apperror.NotFound("todo not found")
	ErrDeletedTodoNotFound = apperror.NotFound("deleted todo not found")
)

type TodoService interface {
	Create(ctx context.Context, userID uint, req *dto.CreateTodoRequest) (*dto.TodoResponse, error)
	GetTodos(ctx context.Context, userID uint, query *dto.TodoQuery) ([]*dto.TodoListResponse, int64, error)
	GetByID(ctx context.Context, userID uint, id uint) (*dto.TodoResponse, error)
	Update(ctx context.Context, userID uint, id uint, req *dto.UpdateTodoRequest) (*dto.TodoResponse, error)
	ToggleComplete(ctx context.Context, userID uint, id uint) (*dto.TodoResponse, error)
	Delete(ctx context.Context, userID uint, id uint) error
	GetDeletedTodos(ctx context.Context, userID uint, query *dto.PaginationQuery) ([]*dto.TodoResponse, int64, error)
	Restore(ctx context.Context, userID uint, id uint) (*dto.TodoResponse, error)
	ForceDelete(ctx context.Context, userID uint, id uint) error
}

type todoService struct {
	todoRepo     repository.TodoRepository
	categoryRepo repository.CategoryRepository
}

func NewTodoService(todoRepo repository.TodoRepository, categoryRepo repository.CategoryRepository) TodoService {
	return &todoService{
		todoRepo:     todoRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *todoService) Create(ctx context.Context, userID uint, req *dto.CreateTodoRequest) (*dto.TodoResponse, error) {
	if req.CategoryID != nil && *req.CategoryID > 0 {
		category, err := s.categoryRepo.FindByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, apperror.Internal(err, "Failed to verify category")
		}
		if category == nil || category.UserID != userID {
			return nil, apperror.BadRequest("Category not found or does not belong to you")
		}
	} else {
		req.CategoryID = nil
	}

	var desc *string
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		trimmed := strings.TrimSpace(*req.Description)
		desc = &trimmed
	}

	todo := &entity.Todo{
		UserID:      userID,
		CategoryID:  req.CategoryID,
		Title:       strings.TrimSpace(req.Title),
		Description: desc,
		IsCompleted: false,
	}

	if err := s.todoRepo.Create(ctx, todo); err != nil {
		return nil, apperror.Internal(err, "Failed to create todo")
	}

	// Re-fetch to preload relations (such as Category)
	if todo.CategoryID != nil {
		if fetched, err := s.todoRepo.FindByID(ctx, todo.ID); err == nil && fetched != nil {
			todo = fetched
		}
	}

	slog.Info("Todo created", "todo_id", todo.ID, "user_id", userID)
	return dto.ToTodoResponse(todo), nil
}

func (s *todoService) GetTodos(ctx context.Context, userID uint, query *dto.TodoQuery) ([]*dto.TodoListResponse, int64, error) {
	page := query.GetPage()
	limit := query.GetLimit()

	todos, total, err := s.todoRepo.FindAllByUserID(ctx, userID, query.Search, query.CategoryID, query.IsCompleted, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "Failed to retrieve todos")
	}

	res := dto.ToTodoListResponse(todos)
	return res, total, nil
}

func (s *todoService) GetByID(ctx context.Context, userID uint, id uint) (*dto.TodoResponse, error) {
	todo, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find todo")
	}
	if todo == nil || todo.UserID != userID {
		return nil, ErrTodoNotFound
	}

	return dto.ToTodoResponse(todo), nil
}

func (s *todoService) Update(ctx context.Context, userID uint, id uint, req *dto.UpdateTodoRequest) (*dto.TodoResponse, error) {
	todo, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find todo")
	}
	if todo == nil || todo.UserID != userID {
		return nil, ErrTodoNotFound
	}

	// 1. Title wajib diisi (validasi dari binding struct, tapi tetap di-trim)
	todo.Title = strings.TrimSpace(req.Title)

	// 2. Description bisa kosong sehingga tersimpan null
	if req.Description == nil || strings.TrimSpace(*req.Description) == "" {
		todo.Description = nil
	} else {
		trimmedDesc := strings.TrimSpace(*req.Description)
		todo.Description = &trimmedDesc
	}

	// 3. CategoryID bisa kosong sehingga tersimpan null
	if req.CategoryID == nil || *req.CategoryID == 0 {
		todo.CategoryID = nil
		todo.Category = nil
	} else {
		category, err := s.categoryRepo.FindByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, apperror.Internal(err, "Failed to verify category")
		}
		if category == nil || category.UserID != userID {
			return nil, apperror.BadRequest("Category not found or does not belong to you")
		}
		todo.CategoryID = req.CategoryID
		todo.Category = category
	}

	// 4. Update status is_completed jika diberikan
	if req.IsCompleted != nil {
		todo.IsCompleted = *req.IsCompleted
	}

	if err := s.todoRepo.Update(ctx, todo); err != nil {
		return nil, apperror.Internal(err, "Failed to update todo")
	}

	if fetched, err := s.todoRepo.FindByID(ctx, todo.ID); err == nil && fetched != nil {
		todo = fetched
	}

	slog.Info("Todo updated", "todo_id", todo.ID, "user_id", userID)
	return dto.ToTodoResponse(todo), nil
}

func (s *todoService) ToggleComplete(ctx context.Context, userID uint, id uint) (*dto.TodoResponse, error) {
	todo, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find todo")
	}
	if todo == nil || todo.UserID != userID {
		return nil, ErrTodoNotFound
	}

	// Pengecekan kondisi field is_completed:
	// Jika belum selesai (false) menjadi selesai (true), dan sebaliknya
	todo.IsCompleted = !todo.IsCompleted

	if err := s.todoRepo.Update(ctx, todo); err != nil {
		return nil, apperror.Internal(err, "Failed to update todo completion status")
	}

	if fetched, err := s.todoRepo.FindByID(ctx, todo.ID); err == nil && fetched != nil {
		todo = fetched
	}

	statusStr := "uncompleted"
	if todo.IsCompleted {
		statusStr = "completed"
	}
	slog.Info("Todo completion status toggled", "todo_id", id, "user_id", userID, "status", statusStr)
	return dto.ToTodoResponse(todo), nil
}

func (s *todoService) Delete(ctx context.Context, userID uint, id uint) error {
	todo, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return apperror.Internal(err, "Failed to find todo")
	}
	if todo == nil || todo.UserID != userID {
		return ErrTodoNotFound
	}

	if err := s.todoRepo.Delete(ctx, id); err != nil {
		return apperror.Internal(err, "Failed to delete todo")
	}

	slog.Info("Todo soft deleted", "todo_id", id, "user_id", userID)
	return nil
}

func (s *todoService) GetDeletedTodos(ctx context.Context, userID uint, query *dto.PaginationQuery) ([]*dto.TodoResponse, int64, error) {
	page := query.GetPage()
	limit := query.GetLimit()

	todos, total, err := s.todoRepo.FindAllDeletedByUserID(ctx, userID, query.Search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "Failed to retrieve deleted todos")
	}

	res := make([]*dto.TodoResponse, 0, len(todos))
	for _, t := range todos {
		res = append(res, dto.ToTodoResponse(t))
	}

	return res, total, nil
}

func (s *todoService) Restore(ctx context.Context, userID uint, id uint) (*dto.TodoResponse, error) {
	todo, err := s.todoRepo.FindDeletedByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find deleted todo")
	}
	if todo == nil || todo.UserID != userID {
		return nil, ErrDeletedTodoNotFound
	}

	if err := s.todoRepo.Restore(ctx, id); err != nil {
		return nil, apperror.Internal(err, "Failed to restore todo")
	}

	restored, err := s.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to retrieve restored todo")
	}

	slog.Info("Todo restored", "todo_id", id, "user_id", userID)
	return dto.ToTodoResponse(restored), nil
}

func (s *todoService) ForceDelete(ctx context.Context, userID uint, id uint) error {
	todo, err := s.todoRepo.FindDeletedByID(ctx, id)
	if err != nil {
		return apperror.Internal(err, "Failed to find deleted todo")
	}
	if todo == nil || todo.UserID != userID {
		return ErrDeletedTodoNotFound
	}

	if err := s.todoRepo.ForceDelete(ctx, id); err != nil {
		return apperror.Internal(err, "Failed to permanently delete todo")
	}

	slog.Info("Todo permanently deleted", "todo_id", id, "user_id", userID)
	return nil
}
