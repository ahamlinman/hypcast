// Package client enables embedding of locally built Hypcast client assets.
package client

import "io/fs"

// Build embeds the Hypcast client when the "embedclient" build tag is set. When
// Build is nil, embedded client assets are not available. When it is not nil,
// it is rooted inside of the "build" directory produced by the build process.
var Build fs.FS
