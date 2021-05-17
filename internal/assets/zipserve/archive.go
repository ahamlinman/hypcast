package zipserve

import (
	"archive/zip"
	"path"
	"strings"
)

// Archive represents a ZIP archive, from which file entries can be found and
// read.
type Archive struct {
	zr *zip.Reader

	entries map[string]*Entry
}

func newArchive(zr *zip.Reader) *Archive {
	a := &Archive{zr: zr}
	a.init()
	return a
}

func (a *Archive) init() {
	a.entries = make(map[string]*Entry)

	// Always synthesize an entry for the root of the ZIP file to treat it as a
	// directory.
	a.entries[cleanPath("/")] = &Entry{}

	// Create real entries for the files in the archive.
	for _, zf := range a.zr.File {
		a.entries[cleanPath(zf.Name)] = &Entry{zf: zf}
	}

	// Finally, synthesize "missing" directory entries for downstream convenience.
	for entrypath := range a.entries {
		dirpath := cleanPath(path.Dir(entrypath))
		if a.entries[dirpath] == nil {
			a.entries[dirpath] = &Entry{}
		}
	}
}

func (a *Archive) Find(path string) *Entry {
	return a.entries[cleanPath(path)]
}

func cleanPath(p string) string {
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	return p
}

type Entry struct {
	// For all real entries in the ZIP file, zf points to the original *zip.Reader
	// entry. Archive will also synthesize directory entries whenever the original
	// ZIP file doesn't contain them (e.g. at the root), in which case zf == nil.
	zf *zip.File
}

func (f *Entry) IsDir() bool {
	if f.zf == nil {
		return true
	}

	return f.zf.Mode().IsDir()
}
