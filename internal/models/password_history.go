package models

import "time"

// PasswordHistory records a previously used password hash for a user.
// Used to enforce password reuse prevention (PCI DSS 8.3.7).
// Records are append-only and never soft-deleted.
type PasswordHistory struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	UserID       uint      `gorm:"not null;index"           json:"-"`
	PasswordHash string    `gorm:"size:255;not null"        json:"-"`
	CreatedAt    time.Time `                                json:"-"`
}
