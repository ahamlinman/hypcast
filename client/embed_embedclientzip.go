// +build embedclientzip

package client

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"io/fs"
)

var (
	//go:embed build.zip
	buildZip    []byte
	buildReader = bytes.NewReader(buildZip)
)

func init() {
	zr, err := zip.NewReader(buildReader, buildReader.Size())
	if err != nil {
		panic(err)
	}

	sub, err := fs.Sub(zr, "build")
	if err != nil {
		panic(err)
	}

	Build = sub
}
