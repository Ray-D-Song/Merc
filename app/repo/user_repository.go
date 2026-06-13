// Package repo encapsulates database operations
package repo

import (
	"context"
	"errors"

	"github.com/ray-d-song/merc/app/infra/dbx"
	"github.com/ray-d-song/merc/app/model"
	"gorm.io/gorm"
)

// UserRepository defines persistence operations for User.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	List(ctx context.Context, query UserQuery, offset, limit int) ([]*model.User, error)
	Count(ctx context.Context, query UserQuery) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id uint, isEnabled bool) error
	UpdateRole(ctx context.Context, id uint, role string, isAdmin bool) error
}

// UserQuery contains filter conditions for querying users
type UserQuery struct {
	Username  string // Partial match search
	Role      string // Exact match search
	IsEnabled *bool  // Enabled status filter (nil = all, true = enabled, false = disabled)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs a UserRepository backed by gorm.DB.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by their ID
func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateStatus updates the enabled status of a user
func (r *userRepository) UpdateStatus(ctx context.Context, id uint, isEnabled bool) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("is_enabled", isEnabled).Error
}

// CountAll returns the total number of all users in the database
func (r *userRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

// UpdateRole updates the role and isAdmin status of a user
func (r *userRepository) UpdateRole(ctx context.Context, id uint, role string, isAdmin bool) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"role":     role,
			"is_admin": isAdmin,
		}).Error
}

// applyUserFilters is a GORM scope that applies user-specific filters
// This scope encapsulates all user filtering logic in one place
func applyUserFilters(query UserQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Apply username filter (partial match)
		if query.Username != "" {
			db = db.Where("username LIKE ?", "%"+query.Username+"%")
		}

		// Apply role filter (exact match)
		if query.Role != "" {
			db = db.Where("role = ?", query.Role)
		}

		// Apply enabled status filter
		if query.IsEnabled != nil {
			db = db.Where("is_enabled = ?", *query.IsEnabled)
		}

		return db
	}
}

// List retrieves a paginated list of users with optional filters
func (r *userRepository) List(ctx context.Context, query UserQuery, offset, limit int) ([]*model.User, error) {
	var users []*model.User

	err := r.db.WithContext(ctx).
		Scopes(
			applyUserFilters(query),
			dbx.Paginate(offset, limit),
		).
		Order("id DESC"). // Sort by ID descending (newest first)
		Find(&users).Error

	return users, err
}

// Count returns the total number of users matching the query filters
func (r *userRepository) Count(ctx context.Context, query UserQuery) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Scopes(applyUserFilters(query)).
		Count(&count).Error

	return count, err
}
