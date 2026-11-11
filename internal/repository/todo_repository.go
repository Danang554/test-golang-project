package repository

import (
	"context"
	"errors"
	"strings"

	"test-golang-project/internal/entity"

	"gorm.io/gorm"
)

type TodoRepository interface {
	Create(ctx context.Context, todo *entity.Todo) error
	FindByID(ctx context.Context, id uint) (*entity.Todo, error)
	FindAllByUserID(ctx context.Context, userID uint, search string, categoryID *uint, isCompleted *bool, page int, limit int) ([]*entity.Todo, int64, error)
	Update(ctx context.Context, todo *entity.Todo) error
	Delete(ctx context.Context, id uint) error
	FindAllDeletedByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Todo, int64, error)
	FindDeletedByID(ctx context.Context, id uint) (*entity.Todo, error)
	Restore(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
}

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo *entity.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

func (r *todoRepository) FindByID(ctx context.Context, id uint) (*entity.Todo, error) {
	var todo entity.Todo
	err := r.db.WithContext(ctx).Preload("Category").First(&todo, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) FindAllByUserID(ctx context.Context, userID uint, search string, categoryID *uint, isCompleted *bool, page int, limit int) ([]*entity.Todo, int64, error) {
	var todos []*entity.Todo
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Todo{}).Where("user_id = ?", userID)

	if search = strings.TrimSpace(search); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if isCompleted != nil {
		query = query.Where("is_completed = ?", *isCompleted)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("Category").Order("id DESC").Offset(offset).Limit(limit).Find(&todos).Error; err != nil {
		return nil, 0, err
	}

	return todos, total, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *entity.Todo) error {
	return r.db.WithContext(ctx).Save(todo).Error
}

func (r *todoRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Todo{}, id).Error
}

func (r *todoRepository) FindAllDeletedByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Todo, int64, error) {
	var todos []*entity.Todo
	var total int64

	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Todo{}).
		Where("user_id = ? AND deleted_at IS NOT NULL", userID)

	if search = strings.TrimSpace(search); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("Category").Order("deleted_at DESC").Offset(offset).Limit(limit).Find(&todos).Error; err != nil {
		return nil, 0, err
	}

	return todos, total, nil
}

func (r *todoRepository) FindDeletedByID(ctx context.Context, id uint) (*entity.Todo, error) {
	var todo entity.Todo
	err := r.db.WithContext(ctx).Unscoped().Preload("Category").
		Where("id = ? AND deleted_at IS NOT NULL", id).First(&todo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) Restore(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().
		Model(&entity.Todo{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil).Error
}

func (r *todoRepository) ForceDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Todo{}, id).Error
}
