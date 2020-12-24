package assets

import (
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
func (fsys FileSystem) Open(name string) (http.File, error) {
	f, err := fsys.FileSystem.Open(name)
	switch {
	case os.IsNotExist(err) && name != indexPage:
		// Treat this as a single page app route, and attempt to serve the root
		// index page.
		return fsys.Open(indexPage)
	case err != nil:
		return nil, err
	}

	if s, _ := f.Stat(); s.IsDir() {
		// Determine whether there is an index page in the directory. If there is,
		// http.FileServer will serve that instead of a directory listing, and we
		// can simply return f. Otherwise, prevent http.FileServer from seeing the
		// raw directory.
		//
		// This code is a combination of the following two strategies:
		// - https://github.com/jordan-wright/unindexed/blob/master/unindexed.go
		// - https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings

		index, err := fsys.FileSystem.Open(filepath.Join(name, indexPage))
		if err != nil {
			f.Close()
			return nil, err
		}
		index.Close()
	}

	return f, nil
}
