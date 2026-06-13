// Package dto describe the structure of requests and responses
package dto

// RegisterRequest validates payload for registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest validates payload for login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse describes user details returned after auth flows.
type AuthResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Realname string `json:"realname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsAdmin  bool   `json:"isAdmin"`
}
