package zipserve

import (
	"archive/zip"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
)

// Handler serves HTTP requests with the contents of a ZIP archive.
//
// When an HTTP client that supports gzip encoding requests an entry that is
// stored with compression in the archive, Handler will serve the compressed
// file stream directly to the client. Otherwise, Handler will transparently
// serve an uncompressed file.
type Handler struct {
	zr    *zip.Reader
	index map[npath]*zip.File
}

// NewHandler returns a handler that serves HTTP requests with the contents of
// the archive represented by zr.
func NewHandler(zr *zip.Reader) *Handler {
	index := make(map[npath]*zip.File)
	for _, f := range zr.File {
		// It is theoretically possible for two archive entries to end up at the
		// same relative path, in which case we let the latest one win. This should
		// be exceptionally rare; I'm not sure where it would show up outside of a
		// specially crafted archive.
		np := normalizePath(f.Name)
		index[np] = f
	}

	return &Handler{
		zr:    zr,
		index: index,
	}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Index page handling, including redirection to canonical paths.
	// TODO: Single page app handling.
	// TODO: Optimized HEAD requests.

	np := normalizePath(r.URL.Path)
	f, ok := h.index[np]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// TODO: Content-based MIME type sniffing.
	if ctype := mime.TypeByExtension(path.Ext(string(np))); ctype != "" {
		w.Header().Add("Content-Type", ctype)
	}

	// TODO: Serve compressed content if possible... obviously.
	fr, err := f.Open()
	if err != nil {
		http.Error(w, "unable to open file", http.StatusInternalServerError)
		return
	}
	defer fr.Close()

	// TODO: Caching headers (If-Modified-Since, ETag).

	w.WriteHeader(http.StatusOK)
	io.Copy(w, fr)
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
