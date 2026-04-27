package models

import "time"

// PasswordHistory records a previously used bcrypt hash for a user.
// It enforces the password-reuse prevention rule from PCI DSS 8.3.7.
// Records are append-only; they are never soft-deleted.
type PasswordHistory struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID       uint      `gorm:"not null;index"           json:"-"`
	PasswordHash string    `gorm:"size:255;not null"        json:"-"`
	CreatedAt    time.Time `                                json:"-"`
}
