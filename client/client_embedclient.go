// +build embedclient

package client

import (
	"embed"
	"io/fs"
)

//go:embed build
var build embed.FS

func init() {
	var err error
	FS, err = fs.Sub(build, "build")
	if err != nil {
		panic(err)
	}
}
