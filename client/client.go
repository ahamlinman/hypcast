// Package client enables embedding of locally built Hypcast client assets.
package client

import "io/fs"

// Build embeds the Hypcast client when the "embedclient" build tag is set.
//
// When Build is nil, embedded client assets are not available. When it is not
// nil, its root is inside the output directory of the client build process,
// such that index.html is at the top level.
var Build fs.FS
