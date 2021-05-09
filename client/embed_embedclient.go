// +build embedclient

package client

import (
	"embed"
	"io/fs"
)

//go:embed build
var buildRoot embed.FS

func init() {
	buildSub, err := fs.Sub(buildRoot, "build")
	if err != nil {
		panic(err)
	}
	Build = buildSub
}
