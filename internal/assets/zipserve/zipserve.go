// Package zipserve efficiently serves assets from a ZIP file to HTTP clients.
package zipserve

import (
	"archive/zip"
	"encoding/binary"
	"io"
	"log"
	"mime"
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
		h.logf("Not found: %s", r.URL.String())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	acceptEncoding := r.Header.Get("Accept-Encoding")
	// TODO: actually parse the header right
	if strings.Contains(acceptEncoding, "gzip") {
		handled := h.serveGzipEncoded(w, file)
		if handled {
			h.logf("Served as gzip: %s", r.URL.String())
			return
		}
	}

	serveUnencoded(w, file)
	h.logf("Served unencoded: %s", r.URL.String())
}

func cleanPath(p string) string {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	return p
}

func (h *Handler) getFileEntry(filePath string) *zip.File {
	// TODO: better way of handling this
	indexPath := cleanPath(path.Join(filePath, "index.html"))

	// TODO: more efficient file location
	for _, file := range h.zr.File {
		zipPath := cleanPath(file.Name)
		if zipPath == filePath || zipPath == indexPath {
			return file
		}
	}

	return nil
}

func (h *Handler) serveGzipEncoded(w http.ResponseWriter, file *zip.File) (handled bool) {
	if file.Method != zip.Deflate {
		return
	}

	offset, err := file.DataOffset()
	if err != nil {
		return
	}

	handled = true
	w.Header().Add("Content-Encoding", "gzip")

	// gzip header for DEFLATE compression
	header := [10]byte{0: 0x1f, 1: 0x8b, 2: 0x08}
	binary.LittleEndian.PutUint32(header[4:8], uint32(file.Modified.Unix()))
	header[9] = 0xff // Unknown OS
	_, err = w.Write(header[:])
	if err != nil {
		return
	}

	sr := io.NewSectionReader(h.r, offset, int64(file.CompressedSize64))
	_, err = io.Copy(w, sr)
	if err != nil {
		return
	}

	var footer [8]byte
	binary.LittleEndian.PutUint32(footer[:4], file.CRC32)
	binary.LittleEndian.PutUint32(footer[4:], uint32(file.UncompressedSize64))
	w.Write(footer[:])

	return
}

func serveUnencoded(w http.ResponseWriter, file *zip.File) {
	f, err := file.Open()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	filePath := cleanPath(file.Name)
	mimeType := mime.TypeByExtension(path.Ext(filePath)) // TODO: content type detection
	w.Header().Set("Content-Type", mimeType)

	// TODO: If-Modified-Since
	// TODO: ETag
	// TODO: Range
	// TODO: HEAD

	w.WriteHeader(http.StatusOK)
	io.Copy(w, f)
}

func (h *Handler) logf(format string, v ...interface{}) {
	joinFmt := "zipserve.Handler(%p): " + format

	joinArgs := make([]interface{}, len(v)+1)
	joinArgs[0] = h
	copy(joinArgs[1:], v)

	log.Printf(joinFmt, joinArgs...)
}
