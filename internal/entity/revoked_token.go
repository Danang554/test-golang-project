package entity

import "time"

type RevokedToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TokenHash string    `gorm:"type:char(64);uniqueIndex;not null" json:"token_hash"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RevokedToken) TableName() string {
	return "revoked_tokens"
}
