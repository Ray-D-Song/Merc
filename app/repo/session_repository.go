package repo

import (
	"context"
	"errors"
	"time"

	"github.com/ray-d-song/merc/app/model"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	FindBySessionIDHash(ctx context.Context, hash string) (*model.Session, error)
	DeleteBySessionIDHash(ctx context.Context, hash string) error
	DeleteByUserID(ctx context.Context, userID uint) error
	DeleteExpired(ctx context.Context, now time.Time) error
	CountActive(ctx context.Context, now time.Time) (int64, error)
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) FindBySessionIDHash(ctx context.Context, hash string) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).Where("session_id_hash = ?", hash).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) DeleteBySessionIDHash(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Where("session_id_hash = ?", hash).Delete(&model.Session{}).Error
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Session{}).Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&model.Session{}).Error
}

func (r *sessionRepository) CountActive(ctx context.Context, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("expires_at > ?", now).
		Count(&count).Error
	return count, err
}
