package model

import "time"

// User represents a registered account stored in the database.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:100" json:"username"`
	Realname     string    `gorm:"size:255" json:"realname"`
	Email        string    `gorm:"size:255" json:"email"`
	Role         string    `gorm:"size:50;default:user" json:"role"`
	IsAdmin      bool      `gorm:"default:false" json:"isAdmin"`
	IsEnabled    bool      `gorm:"default:true;index" json:"isEnabled"` // User account status
	PasswordHash string    `gorm:"size:255" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
