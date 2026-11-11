package entity

import (
	"time"

	"gorm.io/gorm"
)

type Todo struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	User        *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CategoryID  *uint          `json:"category_id" gorm:"index"`
	Category    *Category      `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Description *string        `json:"description" gorm:"type:text"`
	IsCompleted bool           `json:"is_completed" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
