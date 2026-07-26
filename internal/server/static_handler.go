package server

import (
	"io/fs"
	"net/http"
)

// NewStaticHandler serves the compiled UI bundle from the given filesystem
// (in production, the embedded web/dist tree).
func NewStaticHandler(uiFS fs.FS) http.Handler {
	return http.FileServer(http.FS(uiFS))
}
