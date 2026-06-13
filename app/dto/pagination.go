package dto

type PaginationRequest struct {
	Page     int `form:"page" json:"page" binding:"required,min=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"required,min=1,max=100"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// GetOffset calculates the database offset based on page and pageSize
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size as the query limit
func (p *PaginationRequest) GetLimit() int {
	return p.PageSize
}

// CalculateTotalPages computes the total number of pages
func (p *PaginationRequest) CalculateTotalPages(total int64) int {
	if p.PageSize == 0 {
		return 0
	}
	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize > 0 {
		totalPages++
	}
	return totalPages
}
