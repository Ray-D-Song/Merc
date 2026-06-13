package dto

import "time"

type CreateProjectReq struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
}

type UpdateProjectReq struct {
	ID          uint   `json:"id" binding:"required,min=1"`
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Status      uint8  `json:"status"`
}

type ListProjectReq struct {
	PaginationRequest
}

type ProjectResp struct {
	ID                  uint      `json:"id"`
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

type ListProjectResp struct {
	PaginatedResponse[ProjectResp]
}
