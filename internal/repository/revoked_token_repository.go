package repository

import (
	"context"
	"time"

	"test-golang-project/internal/entity"

	"gorm.io/gorm"
)

type RevokedTokenRepository interface {
	Revoke(ctx context.Context, tokenHash string, userID uint, expiresAt time.Time) error
	IsRevoked(ctx context.Context, tokenHash string) (bool, error)
}

type revokedTokenRepository struct {
	db *gorm.DB
}

func NewRevokedTokenRepository(db *gorm.DB) RevokedTokenRepository {
	return &revokedTokenRepository{db: db}
}

func (r *revokedTokenRepository) Revoke(ctx context.Context, tokenHash string, userID uint, expiresAt time.Time) error {
	revokedToken := &entity.RevokedToken{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	return r.db.WithContext(ctx).Create(revokedToken).Error
}

func (r *revokedTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.RevokedToken{}).
		Where("token_hash = ?", tokenHash).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
