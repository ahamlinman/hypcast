package tuner

import (
	"fmt"
	"sync"
	"time"

	"github.com/pion/randutil"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"

	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/gst"
	"github.com/ahamlinman/hypcast/internal/watch"
)

// Status represents the current status of a tuner, which any Client may read as
// necessary.
type Status struct {
	Active  bool
	Channel atsc.Channel

	VideoTrack *webrtc.Track
	AudioTrack *webrtc.Track

	Error error
}

// Tuner represents an ATSC tuner made available to WebRTC clients.
//
// Clients are free to read the current state of the Tuner as necessary in order
// to update their local status. Clients may register with the Tuner to be
// notified when they should reread the current state.
type Tuner struct {
	mu sync.Mutex

	channels   []atsc.Channel
	channelMap map[string]atsc.Channel

	pipeline *gst.Pipeline
	status   *watch.Value
}

// NewTuner creates a new Tuner that can tune to any of the provided channels.
func NewTuner(channels []atsc.Channel) *Tuner {
	var status watch.Value
	status.Set(Status{})

	return &Tuner{
		channels:   channels,
		channelMap: makeChannelMap(channels),
		status:     &status,
	}
}

func makeChannelMap(channels []atsc.Channel) map[string]atsc.Channel {
	m := make(map[string]atsc.Channel)
	for _, c := range channels {
		m[c.Name] = c
	}
	return m
}

// Channels returns the set of channels known to this Tuner.
func (t *Tuner) Channels() []atsc.Channel {
	return t.channels
}

// Status returns the current status of this tuner.
func (t *Tuner) Status() Status {
	return t.status.Get().(Status)
}

// Subscribe sets up a handler function to continuously receive the status of
// the tuner as it is updated, until the associated subscription is canceled.
//
// See the documentation for the watch package for details of how the
// subscription works.
func (t *Tuner) Subscribe(handler func(Status)) *watch.Subscription {
	return t.status.Subscribe(func(x interface{}) {
		handler(x.(Status))
	})
}

// Stop closes any active pipeline for this Tuner, releasing the DVB device.
func (t *Tuner) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	err := t.destroyAnyRunningPipeline()
	t.status.Set(Status{Error: err})
	return err
}

// Tune closes any active pipeline for this Tuner, and starts a new pipeline to
// stream the channel with the provided name.
func (t *Tuner) Tune(channelName string) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	channel, ok := t.channelMap[channelName]
	if !ok {
		return fmt.Errorf("channel %q not available in this tuner", channelName)
	}

	defer func() {
		if err != nil {
			t.destroyAnyRunningPipeline()
			t.status.Set(Status{Error: err})
		}
	}()

	t.destroyAnyRunningPipeline()

	t.pipeline, err = gst.NewPipeline(channel)
	if err != nil {
		return
	}

	status := Status{
		Active:  true,
		Channel: channel,
	}

	streamID := fmt.Sprintf("Tuner(%p)", t)
	status.VideoTrack, status.AudioTrack, err = createTrackPair(streamID)
	if err != nil {
		return
	}

	t.pipeline.SetSink(gst.SinkTypeVideo, sinkTrack(status.VideoTrack, videoClockRate))
	t.pipeline.SetSink(gst.SinkTypeAudio, sinkTrack(status.AudioTrack, audioClockRate))

	err = t.pipeline.Start()
	if err != nil {
		return
	}

	t.status.Set(status)
	return nil
}

func (t *Tuner) destroyAnyRunningPipeline() error {
	defer func() {
		t.pipeline = nil
	}()

	if t.pipeline == nil {
		return nil
	}

	return t.pipeline.Close()
}

const (
	videoClockRate = 90_000
	audioClockRate = 48_000
)

var ssrcGenerator = randutil.NewMathRandomGenerator()

func createTrackPair(streamID string) (video *webrtc.Track, audio *webrtc.Track, err error) {
	video, err = webrtc.NewTrack(
		webrtc.DefaultPayloadTypeVP8, ssrcGenerator.Uint32(), streamID, streamID,
		webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, videoClockRate),
	)
	if err != nil {
		return
	}

	audio, err = webrtc.NewTrack(
		webrtc.DefaultPayloadTypeOpus, ssrcGenerator.Uint32(), streamID, streamID,
		webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, audioClockRate),
	)
	return
}

func sinkTrack(track *webrtc.Track, clockRate int) gst.Sink {
	return gst.Sink(func(b []byte, d time.Duration) {
		track.WriteSample(media.Sample{
			Data:    b,
			Samples: media.NSamples(d, clockRate),
		})
	})
}
