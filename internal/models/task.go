package models

import (
	"time"

	"gorm.io/gorm"
)

// Task represents a work item assigned to a [User].
type Task struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Title       string `gorm:"size:100;not null;uniqueIndex" json:"title"`
	Description string `gorm:"size:255"                     json:"description"`
	Done        bool   `gorm:"default:false"                json:"done"`
	UserID      uint   `gorm:"not null"                     json:"user_id"`
	User        *User  `gorm:"foreignKey:UserID"            json:"user,omitempty"`
}
