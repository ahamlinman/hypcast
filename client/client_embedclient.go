//go:build embedclient && !embedclientzip

package client

import (
	"embed"
	"io/fs"
)

//go:embed dist
var dist embed.FS

func init() {
	if build, err := fs.Sub(dist, "dist"); err == nil {
		Build = build
	} else {
		panic(err)
	}
}
