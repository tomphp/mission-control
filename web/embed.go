// Package webui embeds the compiled TypeScript UI bundle into the Go
// binary so mission-control ships as a single file.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// DistFS returns the compiled UI bundle rooted at its content (so
// dist/index.html is served as /index.html, etc).
func DistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
