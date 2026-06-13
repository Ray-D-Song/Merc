package dto

// UserDTO represents the user data transfer object for API responses
type UserDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Realname  string `json:"realname"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsAdmin   bool   `json:"isAdmin"`
	IsEnabled bool   `json:"isEnabled"`
}

// ListUserRequest contains pagination and filter parameters for listing users
type ListUserRequest struct {
	PaginationRequest
	Username  string `form:"username"`  // Optional: filter by username (partial match)
	Role      string `form:"role"`      // Optional: filter by role (exact match)
	IsEnabled *bool  `form:"isEnabled"` // Optional: filter by enabled status (nil = all, true = enabled, false = disabled)
}

// ListUserResponse wraps the paginated user list response
type ListUserResponse struct {
	PaginatedResponse[UserDTO]
}

// CreateUserRequest contains the data needed to create a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	Realname string `json:"realname" binding:"max=255"`
	Email    string `json:"email" binding:"omitempty,email,max=255"`
	Role     string `json:"role" binding:"required,oneof=user admin"`
}

// CreateUserResponse wraps the created user data
type CreateUserResponse struct {
	UserDTO
}

// UpdateUserStatusRequest contains the user ID for status update operations
type UpdateUserStatusRequest struct {
	UserID uint `json:"userId" binding:"required,min=1"`
}

// UpdateUserRoleRequest contains the data needed to update a user's role
type UpdateUserRoleRequest struct {
	UserID uint   `json:"userId" binding:"required,min=1"`
	Role   string `json:"role" binding:"required,oneof=user admin"`
}
