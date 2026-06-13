package model

import "time"

// Session represents a persisted login session.
type Session struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SessionIDHash string    `gorm:"size:64;not null;uniqueIndex" json:"sessionIdHash"`
	UserID        uint      `gorm:"not null;index" json:"userId"`
	User          User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt     time.Time `json:"createdAt"`
	ExpiresAt     time.Time `gorm:"not null;index" json:"expiresAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
