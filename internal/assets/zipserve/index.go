package zipserve

import (
	"archive/zip"
	"path"
	"strings"
)

type index map[entryPath]entry

func newIndex(files []*zip.File) index {
	idx := make(index)
	for _, f := range files {
		ep := toEntryPath(f.Name)
		idx[ep] = entry{f}
		// TODO: Synthesize directory entries if they don't exist
	}
	return idx
}

type entryPath string

func toEntryPath(name string) entryPath {
	// This is based on the toValidName function of archive/zip/reader.go in Go
	// 1.16.6.
	p := path.Clean(strings.ReplaceAll(name, `\`, `/`))
	p = strings.TrimPrefix(p, "/")
	for strings.HasPrefix(p, "../") {
		p = p[len("../"):]
	}
	return entryPath(p)
}

func (ep entryPath) Dir() entryPath {
	return entryPath(path.Dir(string(ep)))
}

type entry struct {
	file *zip.File
}
