// Package zipserve efficiently serves assets from a ZIP file to HTTP clients.
package zipserve

import (
	"archive/zip"
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
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

	const (
		headerSize = 10
		footerSize = 8
	)
	finalLength := headerSize + file.CompressedSize64 + footerSize

	w.Header().Add("Content-Length", strconv.FormatUint(finalLength, 10))
	w.Header().Add("Content-Encoding", "gzip")

	filePath := cleanPath(file.Name)
	mimeType := mime.TypeByExtension(path.Ext(filePath))
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	} // otherwise, rely on http.ResponseWriter autodetecting the type from written data

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	// Generate the header for a gzip member with DEFLATE compression. See
	// https://datatracker.ietf.org/doc/html/rfc1952 section 2.3.
	const (
		// IDentification 1 and 2; the gzip magic bytes
		id1 = 0x1f
		id2 = 0x8b
		// Compression Method; 8 = "deflate"
		cm = 8
		// FLaGs; we don't bother with fields like name or comment
		flg = 0
		// eXtra FLags; for deflate this indicates whether the compressor used its
		// fastest or slowest algorithm, which ZIP doesn't tell us
		xfl = 0
		// Operating System; 255 = "unknown"
		os = 255
	)
	header := [headerSize]byte{0: id1, 1: id2, 2: cm, 3: flg, 8: xfl, 9: os}
	binary.LittleEndian.PutUint32(header[4:8], uint32(file.Modified.Unix()))
	_, err = bw.Write(header[:])
	if err != nil {
		return
	}

	// The actual compressed blocks, stolen straight out of the ZIP file.
	sr := io.NewSectionReader(h.r, offset, int64(file.CompressedSize64))
	_, err = io.Copy(bw, sr)
	if err != nil {
		return
	}

	// Generate the footer, also section 2.3 of the RFC.
	var footer [footerSize]byte
	// CRC-32, no special magic here.
	binary.LittleEndian.PutUint32(footer[:4], file.CRC32)
	// "Size of the original (uncompressed) input data modulo 2^32." So, I guess
	// if we serve a file more than 4 GiB we're still cool here.
	binary.LittleEndian.PutUint32(footer[4:], uint32(file.UncompressedSize64))
	bw.Write(footer[:])

	return
}

func serveUnencoded(w http.ResponseWriter, file *zip.File) {
	f, err := file.Open()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Length", strconv.FormatUint(file.UncompressedSize64, 10))

	filePath := cleanPath(file.Name)
	mimeType := mime.TypeByExtension(path.Ext(filePath))
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	} // otherwise, rely on http.ResponseWriter autodetecting the type from written data

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
