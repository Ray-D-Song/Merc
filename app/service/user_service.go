package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/repo"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserNotFound indicates the user does not exist
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists indicates duplicate username
	ErrUserAlreadyExists = errors.New("username already exists")
)

type UserService struct {
	repo repo.UserRepository
}

func NewUserService(repo repo.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ListUser retrieves a paginated list of users based on the request filters
// Returns the list of users, total count, and any error encountered
func (s *UserService) ListUser(ctx context.Context, req dto.ListUserRequest) ([]*model.User, int64, error) {
	// Build query conditions from request
	query := repo.UserQuery{
		Username:  req.Username,
		Role:      req.Role,
		IsEnabled: req.IsEnabled,
	}

	// Get total count first
	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// If no records found, return empty list early
	if total == 0 {
		return []*model.User{}, 0, nil
	}

	// Query paginated data
	users, err := s.repo.List(ctx, query, req.GetOffset(), req.GetLimit())
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// CreateUser creates a new user with hashed password
func (s *UserService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*model.User, error) {
	// Check if username already exists
	existing, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user model
	user := &model.User{
		Username:     req.Username,
		Realname:     req.Realname,
		Email:        req.Email,
		Role:         req.Role,
		IsAdmin:      req.Role == "admin",
		IsEnabled:    true, // New users are enabled by default
		PasswordHash: string(hash),
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// DisableUser disables a user account
func (s *UserService) DisableUser(ctx context.Context, userID uint) error {
	// Check if user exists
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update status to disabled
	if err := s.repo.UpdateStatus(ctx, userID, false); err != nil {
		return fmt.Errorf("disable user: %w", err)
	}

	return nil
}

// EnableUser enables a user account
func (s *UserService) EnableUser(ctx context.Context, userID uint) error {
	// Check if user exists
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update status to enabled
	if err := s.repo.UpdateStatus(ctx, userID, true); err != nil {
		return fmt.Errorf("enable user: %w", err)
	}

	return nil
}

// UpdateUserRole updates a user's role
func (s *UserService) UpdateUserRole(ctx context.Context, userID uint, role string) error {
	// Check if user exists
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Determine if the role is admin
	isAdmin := role == "admin"

	// Update role and isAdmin status
	if err := s.repo.UpdateRole(ctx, userID, role, isAdmin); err != nil {
		return fmt.Errorf("update user role: %w", err)
	}

	return nil
}
