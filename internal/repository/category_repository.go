package repository

import (
	"context"
	"errors"
	"strings"

	"test-golang-project/internal/entity"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	FindByID(ctx context.Context, id uint) (*entity.Category, error)
	FindAllByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Category, int64, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uint) error
	FindAllDeletedByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Category, int64, error)
	FindDeletedByID(ctx context.Context, id uint) (*entity.Category, error)
	Restore(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) FindByID(ctx context.Context, id uint) (*entity.Category, error) {
	var category entity.Category
	err := r.db.WithContext(ctx).Preload("User").First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindAllByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Category, int64, error) {
	var categories []*entity.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Category{}).Where("user_id = ?", userID)
	if search = strings.TrimSpace(search); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").Order("id DESC").Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Category{}, id).Error
}

func (r *categoryRepository) FindAllDeletedByUserID(ctx context.Context, userID uint, search string, page int, limit int) ([]*entity.Category, int64, error) {
	var categories []*entity.Category
	var total int64

	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Category{}).
		Where("user_id = ? AND deleted_at IS NOT NULL", userID)
	if search = strings.TrimSpace(search); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").Order("deleted_at DESC").Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) FindDeletedByID(ctx context.Context, id uint) (*entity.Category, error) {
	var category entity.Category
	err := r.db.WithContext(ctx).Unscoped().Preload("User").
		Where("id = ? AND deleted_at IS NOT NULL", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Restore(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().
		Model(&entity.Category{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil).Error
}

func (r *categoryRepository) ForceDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.Category{}, id).Error
}
