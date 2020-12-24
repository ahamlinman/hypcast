package client

import (
	"embed"
	"io/fs"
)

//go:embed build
var build embed.FS

// FS embeds resources for the Hypcast web client.
var FS = must(fs.Sub(build, "build"))

func must(fsys fs.FS, err error) fs.FS {
	if err != nil {
		panic(err)
	}
	return fsys
}
