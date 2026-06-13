// Package model provides binding objects for the database
package model

import "gorm.io/gorm"

// AutoMigrate ensures required tables exist.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Session{}, &Project{})
}
