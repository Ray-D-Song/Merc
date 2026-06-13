package model

import "time"

type Project struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Name                string    `json:"name"`
	Provider            string    `gorm:"size:50;default:github;index" json:"provider"`
	RepositoryURL       string    `gorm:"size:500;uniqueIndex" json:"repositoryUrl"`
	Owner               string    `gorm:"size:255;index" json:"owner"`
	Repo                string    `gorm:"size:255;index" json:"repo"`
	Description         string    `json:"description"`
	Status              uint8     `json:"status"`
	CreaterID           uint      `json:"createrId"`
	CreaterUserName     string    `json:"createrUserName"`
	LastUpdaterID       uint      `json:"lastUpdaterId"`
	LastUpdaterUserName string    `json:"lastUpdaterUserName"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}
