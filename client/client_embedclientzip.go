// +build embedclientzip

package client

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"net/http"

	"github.com/ahamlinman/hypcast/internal/assets"
)

//go:embed build.zip
var buildZip []byte

func init() {
	buildReader := bytes.NewReader(buildZip)
	zr, err := zip.NewReader(buildReader, buildReader.Size())
	if err != nil {
		panic(err)
	}
	Handler = http.FileServer(assets.FileSystem{FileSystem: http.FS(zr)})
}
