// Package zipserve serves assets from a ZIP file to HTTP clients.
package zipserve

import (
	"archive/zip"
	"net/http"

	"github.com/ahamlinman/hypcast/internal/assets"
)

type Handler struct {
	zr *zip.Reader
}

func NewHandler(zr *zip.Reader) *Handler {
	return &Handler{zr}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Replace this with a proper zipserve implementation.
	handler := http.FileServer(assets.FileSystem{FileSystem: http.FS(h.zr)})
	handler.ServeHTTP(w, r)
}
