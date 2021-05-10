// +build embedclientzip

package client

import (
	"bytes"
	_ "embed"

	"github.com/ahamlinman/hypcast/internal/assets/zipserve"
)

//go:embed build.zip
var buildZip []byte

func init() {
	r := bytes.NewReader(buildZip)
	h, err := zipserve.NewHandler(r, r.Size())
	if err != nil {
		panic(err)
	}

	Handler = h
}
