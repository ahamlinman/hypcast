// +build embedclientzip

package client

import (
	"archive/zip"
	"bytes"
	_ "embed"

	"github.com/ahamlinman/hypcast/internal/assets/zipserve"
)

//go:embed build.zip
var buildZip []byte

func init() {
	r := bytes.NewReader(buildZip)
	zr, err := zip.NewReader(r, r.Size())
	if err != nil {
		panic(err)
	}
	Handler = zipserve.NewHandler(zr)
}
