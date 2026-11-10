package entity

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	User        *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
