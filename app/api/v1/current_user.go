package v1

import (
	"net/http"

	"github.com/ray-d-song/merc/app/infra/httpx"
	"github.com/ray-d-song/merc/app/service"
)

func currentUserIdentity(w http.ResponseWriter, r *http.Request, auth *service.AuthService) (uint, string, bool) {
	userID, err := httpx.GetUserID(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return 0, "", false
	}
	user, err := auth.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return 0, "", false
	}
	return user.ID, user.Username, true
}
