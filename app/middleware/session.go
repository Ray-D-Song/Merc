package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/ray-d-song/merc/app/infra/httpx"
	"github.com/ray-d-song/merc/app/store"
)

const (
	// SessionCookieName is the cookie key we use to keep track of the logged-in user.
	SessionCookieName = "session_id"
)

// RequireSession ensures the request contains a valid session cookie; otherwise it aborts with 401.
func RequireSession(sessionStore *store.SessionStore, sessionSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			signedSessionID := cookie.Value

			// Verify signature and extract original session ID
			sessionID, valid := verifySessionID(signedSessionID, sessionSecret)
			if !valid {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			// Validate session against persisted store
			userID, exists, err := sessionStore.Get(r.Context(), sessionID)
			if err != nil {
				httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to validate session"})
				return
			}
			if !exists {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ctx := httpx.SetUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// verifySessionID verifies the signed session ID and returns the original session ID
func verifySessionID(signedSessionID, secret string) (string, bool) {
	parts := strings.Split(signedSessionID, ".")
	if len(parts) != 2 {
		return "", false
	}

	sessionID := parts[0]
	providedSignature := parts[1]

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sessionID))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(expectedSignature), []byte(providedSignature)) {
		return "", false
	}

	return sessionID, true
}
