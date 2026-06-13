package v1

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/infra/config"
	"github.com/ray-d-song/merc/app/infra/httpx"
	"github.com/ray-d-song/merc/app/middleware"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/service"
	"github.com/ray-d-song/merc/app/store"
)

// AuthHandler exposes HTTP handlers for authentication endpoints.
type AuthHandler struct {
	service              *service.AuthService
	sessionStore         *store.SessionStore
	sessionMaxAgeSeconds int
	sessionSecret        string
	cookieSecure         bool
	cookieSameSite       http.SameSite
}

// NewAuthHandler constructs AuthHandler.
func NewAuthHandler(service *service.AuthService, sessionStore *store.SessionStore, cfg *config.AppConfig) *AuthHandler {
	maxAge := int((time.Duration(cfg.Session.MaxAgeMS) * time.Millisecond).Seconds())
	if maxAge <= 0 {
		maxAge = int((14 * 24 * time.Hour) / time.Second)
	}

	return &AuthHandler{
		service:              service,
		sessionStore:         sessionStore,
		sessionMaxAgeSeconds: maxAge,
		sessionSecret:        cfg.Session.Secret,
		cookieSecure:         cfg.Session.CookieSecure,
		cookieSameSite:       parseCookieSameSite(cfg.Session.CookieSameSite),
	}
}

// RegisterRoutes wires /auth endpoints under /api/v1.
func (h *AuthHandler) RegisterRoutes(router chi.Router, sessionStore *store.SessionStore, sessionSecret string) {
	router.Post("/register", h.handleRegister)
	router.Post("/login", h.handleLogin)
	router.Post("/logout", h.handleLogout)

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.RequireSession(sessionStore, sessionSecret))
		r.Get("/me", h.handleGetCurrentUser)
	})
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.service.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrUsernameTaken:
			httpx.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register user"})
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toAuthResponse(user))
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to login"})
		}
		return
	}

	// Generate random session ID and persist the session
	sessionID, err := h.sessionStore.Create(r.Context(), user.ID)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	// Sign the session ID with HMAC-SHA256
	signedSessionID := h.signSessionID(sessionID)

	maxAge := h.sessionMaxAgeSeconds
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    signedSessionID,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.cookieSameSite,
	})

	httpx.WriteJSON(w, http.StatusOK, toAuthResponse(user))
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err == nil && cookie.Value != "" {
		// Verify and extract the original session ID
		if sessionID, valid := h.verifySessionID(cookie.Value); valid {
			// Delete session from store
			_ = h.sessionStore.Delete(r.Context(), sessionID)
		}
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		MaxAge:   -1, // Expire immediately
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.cookieSameSite,
	})

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func parseCookieSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "default":
		return http.SameSiteDefaultMode
	default:
		return http.SameSiteLaxMode
	}
}

func (h *AuthHandler) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by RequireSession middleware)
	userID, err := httpx.GetUserID(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Get user from service
	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
		return
	}

	if user == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAuthResponse(user))
}

// signSessionID signs the session ID using HMAC-SHA256
// Format: sessionID.signature
func (h *AuthHandler) signSessionID(sessionID string) string {
	mac := hmac.New(sha256.New, []byte(h.sessionSecret))
	mac.Write([]byte(sessionID))
	signature := hex.EncodeToString(mac.Sum(nil))
	return sessionID + "." + signature
}

// verifySessionID verifies the signed session ID and returns the original session ID
func (h *AuthHandler) verifySessionID(signedSessionID string) (string, bool) {
	parts := strings.Split(signedSessionID, ".")
	if len(parts) != 2 {
		return "", false
	}

	sessionID := parts[0]
	providedSignature := parts[1]

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(h.sessionSecret))
	mac.Write([]byte(sessionID))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(expectedSignature), []byte(providedSignature)) {
		return "", false
	}

	return sessionID, true
}

func toAuthResponse(user *model.User) dto.AuthResponse {
	return dto.AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Realname: user.Realname,
		Email:    user.Email,
		Role:     user.Role,
		IsAdmin:  user.IsAdmin,
	}
}
