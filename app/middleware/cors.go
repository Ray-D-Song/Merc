package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

// CORS allows credentialed browser requests from configured frontend origins.
func CORS(frontendDomain string) func(http.Handler) http.Handler {
	allowedOrigins := parseAllowedOrigins(frontendDomain)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := origin != "" && isAllowedOrigin(origin, allowedOrigins)
			if allowed {
				header := w.Header()
				header.Set("Access-Control-Allow-Origin", origin)
				header.Set("Access-Control-Allow-Credentials", "true")
				header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				header.Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Requested-With")
				header.Set("Access-Control-Max-Age", "86400")
				header.Add("Vary", "Origin")
				header.Add("Vary", "Access-Control-Request-Method")
				header.Add("Vary", "Access-Control-Request-Headers")
			}

			if allowed && r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parseAllowedOrigins(frontendDomain string) map[string]struct{} {
	origins := make(map[string]struct{})
	for _, item := range strings.Split(frontendDomain, ",") {
		origin := normalizeOrigin(item)
		if origin != "" {
			origins[origin] = struct{}{}
		}
	}
	return origins
}

func isAllowedOrigin(origin string, allowedOrigins map[string]struct{}) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	_, ok := allowedOrigins[normalizeOrigin(origin)]
	return ok
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimRight(origin, "/")
	}

	return parsed.Scheme + "://" + parsed.Host
}
