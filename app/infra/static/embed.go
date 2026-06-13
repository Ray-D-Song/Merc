package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:web-dist
var webDistFS embed.FS

// GetFS returns the embedded filesystem containing the web-dist directory.
// It strips the "web-dist" prefix so files are served from root.
func GetFS() (http.FileSystem, error) {
	stripped, err := fs.Sub(webDistFS, "web-dist")
	if err != nil {
		return nil, err
	}
	return http.FS(stripped), nil
}

// MustGetFS is like GetFS but panics on error. Suitable for initialization.
func MustGetFS() http.FileSystem {
	fsys, err := GetFS()
	if err != nil {
		panic("failed to get embedded filesystem: " + err.Error())
	}
	return fsys
}
