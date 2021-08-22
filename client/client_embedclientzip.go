//go:build embedclientzip

package client

import (
	"archive/zip"
	"bytes"
	_ "embed"
)

//go:embed build.zip
var buildZip []byte

func init() {
	buildReader := bytes.NewReader(buildZip)
	zr, err := zip.NewReader(buildReader, buildReader.Size())
	if err != nil {
		panic(err)
	}
	Build = zr
}
