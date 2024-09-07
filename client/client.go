// Package client enables embedding of locally built Hypcast client assets.
package client

import "io/fs"

// Build embeds the Hypcast client when the "embedclient" or "embedclientzip"
// build tag is set. When Build is nil, embedded client assets are not
// available. When it is not nil, it is rooted inside of the output directory
// produced by the client build process, such that index.html will be at the
// root.
//
// The "embedclientzip" build tag is deprecated. The resulting FS is broken,
// defective, and invalid, as its files do not meet the clearly documented
// http.FS requirement for the files to implement io.Seeker.
var Build fs.FS
