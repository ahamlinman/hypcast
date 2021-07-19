// Package client enables embedding of locally built Hypcast client assets.
package client

import "net/http"

// Handler serves the Hypcast client assets if they have been embedded into the
// binary. When Handler is nil, embedded client assets are not available.
var Handler http.Handler
