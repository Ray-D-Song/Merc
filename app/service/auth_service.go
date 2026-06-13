// Package service is the specific implementation of the service
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/repo"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUsernameTaken indicates a duplicate registration attempt.
	ErrUsernameTaken = errors.New("username already registered")
	// ErrInvalidCredentials indicates bad login credentials.
	ErrInvalidCredentials = errors.New("invalid username or password")
)

// AuthService encapsulates registration and login flows.
type AuthService struct {
	repo repo.UserRepository
}

// NewAuthService constructs AuthService.
func NewAuthService(repo repo.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register creates a new user with hashed password.
// The first user to register automatically becomes an admin.
func (s *AuthService) Register(ctx context.Context, username, password string) (*model.User, error) {
	existing, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Check if this is the first user
	count, err := s.repo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	// First user becomes admin
	role := "user"
	isAdmin := false
	if count == 0 {
		role = "admin"
		isAdmin = true
	}

	user := &model.User{
		Username:     username,
		Role:         role,
		IsAdmin:      isAdmin,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// Login verifies credentials and returns the matching user.
func (s *AuthService) Login(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func (s *AuthService) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}
