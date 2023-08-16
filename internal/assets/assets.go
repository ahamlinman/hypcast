// Package assets provides special functionality for serving the Hypcast web
// frontend.
package assets

import (
	"errors"
	"io/fs"
	"net/http"
	"os"

	"github.com/lpar/gzipped/v2"
)

// Handler returns an http.Handler that supports Hypcast's special asset serving
// functionality. (TODO: Explain what exactly that functionality is.)
func Handler(fsys fs.FS) http.Handler {
	return gzipped.FileServer(gzipped.FS(SPA{fsys}))
}

// SPA wraps an fs.FS such that any attempt to open any possible file path
// actually opens the contents of "index.html" in the wrapped file system.
type SPA struct {
	fs.FS
}

const indexPage = "index.html"

// Open implements fs.FS.
func (fs SPA) Open(name string) (fs.File, error) {
	file, err := fs.FS.Open(name)
	if errors.Is(err, os.ErrNotExist) && name != indexPage {
		return fs.Open(indexPage)
	}
	return file, err
}
