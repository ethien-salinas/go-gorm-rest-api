package models

import "gorm.io/gorm"

type Task struct {
	gorm.Model

	Title       string `gorm:"size:100;not null;uniqueIndex" json:"title"`
	Description string `gorm:"size:255" json:"description"`
	Done        bool   `gorm:"default:false" json:"done"`
	UserID      uint   `gorm:"not null" json:"user_id"`
	User        User   `gorm:"foreignKey:UserID" json:"user"`
}
