package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/repo"
)

// SessionStore manages persisted sessions with automatic expiration.
type SessionStore struct {
	repo   repo.SessionRepository
	maxAge time.Duration
}

// NewSessionStore creates a new database-backed session store.
func NewSessionStore(sessionRepo repo.SessionRepository, maxAge time.Duration) *SessionStore {
	store := &SessionStore{
		repo:   sessionRepo,
		maxAge: maxAge,
	}

	// Start background cleanup goroutine
	go store.cleanupExpiredSessions()

	return store
}

// Create generates a new random session ID and stores the session.
func (s *SessionStore) Create(ctx context.Context, userID uint) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	session := &model.Session{
		SessionIDHash: hashSessionID(sessionID),
		UserID:        userID,
		CreatedAt:     now,
		ExpiresAt:     now.Add(s.maxAge),
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		return "", err
	}

	return sessionID, nil
}

// Get retrieves a session by ID and validates expiration.
func (s *SessionStore) Get(ctx context.Context, sessionID string) (uint, bool, error) {
	sessionIDHash := hashSessionID(sessionID)
	session, err := s.repo.FindBySessionIDHash(ctx, sessionIDHash)
	if err != nil {
		return 0, false, err
	}
	if session == nil {
		return 0, false, nil
	}

	if time.Now().After(session.ExpiresAt) {
		if err := s.repo.DeleteBySessionIDHash(ctx, sessionIDHash); err != nil {
			return 0, false, err
		}
		return 0, false, nil
	}

	return session.UserID, true, nil
}

// Delete removes a session from the store.
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.repo.DeleteBySessionIDHash(ctx, hashSessionID(sessionID))
}

// DeleteByUserID removes all sessions for a specific user.
func (s *SessionStore) DeleteByUserID(ctx context.Context, userID uint) error {
	return s.repo.DeleteByUserID(ctx, userID)
}

// Count returns the total number of active sessions.
func (s *SessionStore) Count(ctx context.Context) (int, error) {
	count, err := s.repo.CountActive(ctx, time.Now())
	return int(count), err
}

// cleanupExpiredSessions runs periodically to remove expired sessions.
func (s *SessionStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		_ = s.repo.DeleteExpired(context.Background(), time.Now())
	}
}

// generateSessionID creates a cryptographically secure random session ID.
func generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashSessionID(sessionID string) string {
	sum := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(sum[:])
}
