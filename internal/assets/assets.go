package assets

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
)

const indexPage = "/index.html"

// FileSystem wraps http.FileSystem to provide two extensions.
//
// First, the root index page is served in place of a 404 for unknown paths.
// This supports single page applications that use client-side routing.
//
// Second, directory listings are not served.
type FileSystem struct {
	http.FileSystem
}

// Open implements http.FileSystem.
func (fs FileSystem) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if errors.Is(err, os.ErrNotExist) && name != indexPage {
		// Treat this as a single page app route, and attempt to serve the root
		// index page.
		return fs.Open(indexPage)
	}
	if err != nil {
		return nil, err
	}

	if s, _ := f.Stat(); s.IsDir() {
		// If the directory contains an index page, http.FileServer will serve it
		// instead of a directory listing, and we can return the directory entry
		// that we opened. Otherwise, we should return an error to block clients
		// from seeing the listing.
		//
		// This code is a combination of the following two strategies:
		// - https://github.com/jordan-wright/unindexed/blob/master/unindexed.go
		// - https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
		//
		// Note that even if we open an index page, we have to return the directory
		// entry and not "short circuit" to return the file entry for the page
		// itself. http.FileServer includes redirection logic around the index page
		// (see its docs), which can send the client into an infinite redirect loop
		// if we lie to it about what's really at the requested path.
		index, err := fs.FileSystem.Open(filepath.Join(name, indexPage))
		if err != nil {
			f.Close()
			return nil, err
		}
		index.Close()
	}

	return f, nil
}
