package middleware

import (
	"net/http"
	"path/filepath"
	"strings"
)

// GzipStaticHandler wraps an http.Handler to serve pre-compressed .gz files
// when available and the client supports gzip encoding.
func GzipStaticHandler(next http.Handler, fileSystem http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Try to find the .gz version of the requested file
		gzipPath := r.URL.Path + ".gz"

		// Check if .gz file exists
		f, err := fileSystem.Open(gzipPath)
		if err != nil {
			// No .gz file, serve normally
			next.ServeHTTP(w, r)
			return
		}
		defer f.Close()

		// Verify it's a file, not a directory
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			next.ServeHTTP(w, r)
			return
		}

		// Set Content-Encoding header to indicate gzip
		w.Header().Set("Content-Encoding", "gzip")

		// Set appropriate Content-Type based on file extension
		contentType := getContentType(r.URL.Path)
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		// Vary header to indicate response varies based on Accept-Encoding
		w.Header().Set("Vary", "Accept-Encoding")

		// Serve the .gz file
		r.URL.Path = gzipPath
		next.ServeHTTP(w, r)
	})
}

// getContentType returns the appropriate content type based on file extension
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".ico":
		return "image/x-icon"
	default:
		return ""
	}
}
