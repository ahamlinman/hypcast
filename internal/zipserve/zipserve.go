package zipserve

import (
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

// Handler serves HTTP requests with the contents of a filesystem, serving
// pre-compressed content when available to clients that support it.
type Handler struct {
	fsys fs.FS
}

// NewHandler returns a handler that serves HTTP requests with the contents of
// fsys.
func NewHandler(fsys fs.FS) *Handler {
	return &Handler{fsys}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Index page handling, including redirection to canonical paths.
	// TODO: Single page app handling.
	// TODO: Optimized HEAD requests.

	np := normalizePath(r.URL.Path)
	f, err := h.fsys.Open(string(np))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		http.Error(w, "not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "error opening file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "error reading file info", http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// TODO: Content-based MIME type sniffing.
	if ctype := mime.TypeByExtension(path.Ext(string(np))); ctype != "" {
		w.Header().Add("Content-Type", ctype)
	}

	// TODO: Caching headers (If-Modified-Since, ETag).

	w.WriteHeader(http.StatusOK)
	io.Copy(w, f)
}

// npath holds a normalized path: a Clean relative path separated by forward
// slashes. Because an npath is relative, it contains no leading or trailing
// slashes, and the root path is represented as ".".
type npath string

// normalizePath converts an arbitrary slash-separated path to an npath.
func normalizePath(p string) npath {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		p = "."
	}
	return npath(p)
}
