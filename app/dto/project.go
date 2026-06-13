package dto

import "time"

type CreateProjectReq struct {
	RepositoryURL string `json:"repositoryUrl" binding:"required,min=1,max=500"`
	Name          string `json:"name" binding:"max=255"`
	Description   string `json:"description" binding:"max=1000"`
}

type UpdateProjectReq struct {
	ID            uint   `json:"id" binding:"required,min=1"`
	RepositoryURL string `json:"repositoryUrl" binding:"max=500"`
	Name          string `json:"name" binding:"required,min=1,max=255"`
	Description   string `json:"description" binding:"max=1000"`
	Status        uint8  `json:"status"`
}

type ListProjectReq struct {
	PaginationRequest
}

type ProjectResp struct {
	ID                  uint      `json:"id"`
	Name                string    `json:"name"`
	Provider            string    `json:"provider"`
	RepositoryURL       string    `json:"repositoryUrl"`
	Owner               string    `json:"owner"`
	Repo                string    `json:"repo"`
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
