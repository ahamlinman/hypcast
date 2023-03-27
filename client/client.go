// Package client enables embedding of locally built Hypcast client assets.
package client

import "io/fs"

// Build embeds the Hypcast client when the "embedclientzip" build tag is set.
// When Build is nil, embedded client assets are not available. When it is not
// nil, it is rooted inside of the output directory produced by the client build
// process, such that index.html will be at the root.
var Build fs.FS
