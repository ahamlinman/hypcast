// Package zipserve efficiently serves assets from a ZIP file to HTTP clients.
package zipserve

import (
	"archive/zip"
	"io"
	"net/http"
	"path"
	"strings"
)

// Handler serves assets from a ZIP file to HTTP clients.
//
// When a client indicates support for compression through the Accept-Encoding
// header, Handler will attempt to serve compressed data streams from the ZIP
// file directly to the client, without compressing or decompressing data on the
// server. Handler will transparently decompress file contents for clients that
// do not support compressed transfers.
type Handler struct {
	r  io.ReaderAt
	zr *zip.Reader
}

// NewHandler creates a Handler that serves from the ZIP file read from r with
// the provided size.
func NewHandler(r io.ReaderAt, size int64) (*Handler, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	return &Handler{
		r:  r,
		zr: zr,
	}, nil
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath := cleanPath(r.URL.Path)
	file := h.getFileEntry(filePath)
	if file == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// TODO: serve encoded versions

	serveUnencoded(w, file)
}

func cleanPath(p string) string {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	return p
}

func (h *Handler) getFileEntry(path string) *zip.File {
	// TODO: more efficient file location
	for _, file := range h.zr.File {
		filePath := cleanPath(file.Name)
		if path == filePath {
			return file
		}
	}

	return nil
}

func serveUnencoded(w http.ResponseWriter, file *zip.File) {
	f, err := file.Open()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.WriteHeader(http.StatusOK)
	io.Copy(w, f)
}
