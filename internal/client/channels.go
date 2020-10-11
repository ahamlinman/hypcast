package client

import (
	"encoding/json"
	"net/http"

	"github.com/ahamlinman/hypcast/internal/atsc"
)

// ChannelListHandler returns a http.Handler that responds to requests with a
// JSON list of available channel names.
func ChannelListHandler(channels []atsc.Channel) http.Handler {
	channelNames := make([]string, len(channels))
	for i, ch := range channels {
		channelNames[i] = ch.Name
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(channelNames)
	})
}
