package middleware

import (
	"net/http"

	"github.com/ray-d-song/merc/app/infra/httpx"
	"github.com/ray-d-song/merc/app/repo"
)

// RequireAdmin is a middleware that checks if the current user is an admin.
// It must be used after RequireSession middleware.
func RequireAdmin(userRepo repo.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by RequireSession middleware)
			userID, err := httpx.GetUserID(r)
			if err != nil {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			// Fetch user from database
			user, err := userRepo.FindByID(r.Context(), userID)
			if err != nil || user == nil {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			// Check if user is admin
			if !user.IsAdmin {
				httpx.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
