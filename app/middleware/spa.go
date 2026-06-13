package middleware

import (
	"net/http"
	"path/filepath"
	"strings"
)

// SPAHandler serves static files and falls back to index.html for SPA routes.
// It checks if the requested file exists; if not, serves index.html instead.
// It also supports serving pre-compressed .gz files when available.
func SPAHandler(fileSystem http.FileSystem) http.Handler {
	fileServer := http.FileServer(fileSystem)

	// Wrap with gzip static handler to serve pre-compressed files
	gzipHandler := GzipStaticHandler(fileServer, fileSystem)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip API routes - they're handled by other routes
		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to open the file (check both original and .gz)
		f, err := fileSystem.Open(path)
		fileExists := false

		if err != nil {
			// Try .gz version
			fGz, errGz := fileSystem.Open(path + ".gz")
			if errGz == nil {
				fGz.Close()
				fileExists = true
			}
		} else {
			stat, err := f.Stat()
			f.Close()

			if err == nil && stat.IsDir() {
				// It's a directory, check for index.html or index.html.gz
				indexPath := filepath.Join(path, "index.html")
				_, errIndex := fileSystem.Open(indexPath)
				_, errIndexGz := fileSystem.Open(indexPath + ".gz")

				if errIndex == nil || errIndexGz == nil {
					// index.html or index.html.gz exists, serve it
					r.URL.Path = indexPath
					fileExists = true
				}
			} else if err == nil {
				// It's a file and exists
				fileExists = true
			}
		}

		// If neither original nor .gz exists, fallback to /index.html for SPA routing
		if !fileExists {
			r.URL.Path = "/index.html"
		}

		gzipHandler.ServeHTTP(w, r)
	})
}
