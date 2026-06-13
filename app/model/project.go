package model

import "time"

type Project struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Status              uint8     `json:"status"`
	CreaterID           uint      `json:"createrId"`
	CreaterUserName     string    `json:"createrUserName"`
	LastUpdaterID       uint      `json:"lastUpdaterId"`
	LastUpdaterUserName string    `json:"lastUpdaterUserName"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}
