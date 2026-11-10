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
	ErrCategoryNotFound        = apperror.NotFound("category not found")
	ErrDeletedCategoryNotFound = apperror.NotFound("deleted category not found")
)

type CategoryService interface {
	Create(ctx context.Context, userID uint, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetCategories(ctx context.Context, userID uint, query *dto.PaginationQuery) ([]*dto.CategoryListResponse, int64, error)
	GetByID(ctx context.Context, userID uint, id uint) (*dto.CategoryResponse, error)
	Update(ctx context.Context, userID uint, id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	Delete(ctx context.Context, userID uint, id uint) error
	GetDeletedCategories(ctx context.Context, userID uint, query *dto.PaginationQuery) ([]*dto.CategoryResponse, int64, error)
	Restore(ctx context.Context, userID uint, id uint) (*dto.CategoryResponse, error)
	ForceDelete(ctx context.Context, userID uint, id uint) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: categoryRepo}
}

func (s *categoryService) Create(ctx context.Context, userID uint, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	category := &entity.Category{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, apperror.Internal(err, "Failed to create category")
	}

	slog.Info("Category created", "category_id", category.ID, "user_id", userID)
	return dto.ToCategoryResponse(category), nil
}

func (s *categoryService) GetCategories(ctx context.Context, userID uint, query *dto.PaginationQuery) ([]*dto.CategoryListResponse, int64, error) {
	page := query.GetPage()
	limit := query.GetLimit()

	categories, total, err := s.categoryRepo.FindAllByUserID(ctx, userID, query.Search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "Failed to retrieve categories")
	}

	return dto.ToCategoryListResponse(categories), total, nil
}

func (s *categoryService) GetByID(ctx context.Context, userID uint, id uint) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to retrieve category")
	}
	if category == nil || category.UserID != userID {
		return nil, ErrCategoryNotFound
	}

	return dto.ToCategoryResponse(category), nil
}

func (s *categoryService) Update(ctx context.Context, userID uint, id uint, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find category")
	}
	if category == nil || category.UserID != userID {
		return nil, ErrCategoryNotFound
	}

	hasChanges := false
	if strings.TrimSpace(req.Name) != "" {
		category.Name = strings.TrimSpace(req.Name)
		hasChanges = true
	}
	if strings.TrimSpace(req.Description) != "" {
		category.Description = strings.TrimSpace(req.Description)
		hasChanges = true
	}

	if hasChanges {
		if err := s.categoryRepo.Update(ctx, category); err != nil {
			return nil, apperror.Internal(err, "Failed to update category")
		}
		slog.Info("Category updated", "category_id", category.ID, "user_id", userID)
	}

	return dto.ToCategoryResponse(category), nil
}

func (s *categoryService) Delete(ctx context.Context, userID uint, id uint) error {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return apperror.Internal(err, "Failed to find category")
	}
	if category == nil || category.UserID != userID {
		return ErrCategoryNotFound
	}

	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return apperror.Internal(err, "Failed to delete category")
	}

	slog.Info("Category soft deleted", "category_id", id, "user_id", userID)
	return nil
}

func (s *categoryService) GetDeletedCategories(ctx context.Context, userID uint, query *dto.PaginationQuery) ([]*dto.CategoryResponse, int64, error) {
	page := query.GetPage()
	limit := query.GetLimit()

	categories, total, err := s.categoryRepo.FindAllDeletedByUserID(ctx, userID, query.Search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "Failed to retrieve deleted categories")
	}

	res := make([]*dto.CategoryResponse, 0, len(categories))
	for _, c := range categories {
		res = append(res, dto.ToCategoryResponse(c))
	}

	return res, total, nil
}

func (s *categoryService) Restore(ctx context.Context, userID uint, id uint) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindDeletedByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to find deleted category")
	}
	if category == nil || category.UserID != userID {
		return nil, ErrDeletedCategoryNotFound
	}

	if err := s.categoryRepo.Restore(ctx, id); err != nil {
		return nil, apperror.Internal(err, "Failed to restore category")
	}

	restored, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "Failed to retrieve restored category")
	}

	slog.Info("Category restored", "category_id", id, "user_id", userID)
	return dto.ToCategoryResponse(restored), nil
}

func (s *categoryService) ForceDelete(ctx context.Context, userID uint, id uint) error {
	category, err := s.categoryRepo.FindDeletedByID(ctx, id)
	if err != nil {
		return apperror.Internal(err, "Failed to find deleted category")
	}
	if category == nil || category.UserID != userID {
		return ErrDeletedCategoryNotFound
	}

	if err := s.categoryRepo.ForceDelete(ctx, id); err != nil {
		return apperror.Internal(err, "Failed to permanently delete category")
	}

	slog.Info("Category permanently deleted", "category_id", id, "user_id", userID)
	return nil
}
