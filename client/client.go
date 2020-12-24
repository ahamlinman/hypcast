package client

import "io/fs"

// FS embeds resources for the Hypcast web client when the "embedclient" build
// tag is set. FS will be nil when client resources are not available.
var FS fs.FS
