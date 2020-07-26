package tuner

import (
	"sync"

	"github.com/pion/webrtc/v2"

	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/gst"
)

type Client interface {
	SendAvailableChannels([]atsc.Channel)
	SendStatus(active bool, channel atsc.Channel)
	SendVideoTrack(*webrtc.Track)
	SendAudioTrack(*webrtc.Track)
}

type Tuner struct {
	channels []atsc.Channel

	active     bool
	channel    atsc.Channel
	pipeline   *gst.Pipeline
	videoTrack *webrtc.Track
	audioTrack *webrtc.Track

	clientsMu sync.Mutex
	clients   map[Client]struct{}
}

func NewTuner(channels []atsc.Channel) *Tuner {
	return &Tuner{
		channels: channels,
		clients:  make(map[Client]struct{}),
	}
}
