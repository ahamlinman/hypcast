package tuner

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pion/randutil"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"

	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/gst"
	"github.com/ahamlinman/hypcast/internal/watch"
)

// State represents the current state of the tuner.
type State int

const (
	// StateStopped means the tuner is switched off.
	StateStopped State = iota
	// StateStarting means that the tuner is trying to lock onto a signal and
	// start streaming.
	StateStarting
	// StatePlaying means that the tuner is locked onto a singal and is actively
	// streaming video.
	StatePlaying
)

// Status represents the public state of the tuner for reading by clients.
type Status struct {
	State   State
	Channel atsc.Channel
	Error   error
}

// Tracks represents the current set of video and audio tracks for use by WebRTC
// clients.
type Tracks struct {
	Video *webrtc.Track
	Audio *webrtc.Track
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
	tracks   *watch.Value
}

// NewTuner creates a new Tuner that can tune to any of the provided channels.
func NewTuner(channels []atsc.Channel) *Tuner {
	return &Tuner{
		channels:   channels,
		channelMap: makeChannelMap(channels),
		status:     watch.NewValue(Status{}),
		tracks:     watch.NewValue(Tracks{}),
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
	channels := make([]atsc.Channel, len(t.channels))
	copy(channels, t.channels)
	return channels
}

// SubscribeStatus sets up a handler function to continuously receive the status
// of the tuner as it is updated, until the associated subscription is canceled.
//
// See the documentation for the watch package for details of how the
// subscription works.
func (t *Tuner) SubscribeStatus(handler func(Status)) *watch.Subscription {
	return t.status.Subscribe(func(x interface{}) {
		handler(x.(Status))
	})
}

// SubscribeTracks sets up a handler function to continuously receive the
// tuner's WebRTC tracks as they are updated, until the associated subscription
// is canceled.
//
// See the documentation for the watch package for details of how the
// subscription works.
func (t *Tuner) SubscribeTracks(handler func(Tracks)) *watch.Subscription {
	return t.tracks.Subscribe(func(x interface{}) {
		handler(x.(Tracks))
	})
}

// Stop closes any active pipeline for this Tuner, releasing the DVB device.
func (t *Tuner) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	err := t.destroyAnyRunningPipeline()
	t.status.Set(Status{Error: err})
	t.tracks.Set(Tracks{})
	return err
}

// ErrChannelNotFound is returned when tuning to a channel that does not exist.
var ErrChannelNotFound error = errors.New("channel not found")

// Tune closes any active pipeline for this Tuner, and starts a new pipeline to
// stream the channel with the provided name.
func (t *Tuner) Tune(channelName string) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	channel, ok := t.channelMap[channelName]
	if !ok {
		return ErrChannelNotFound
	}

	t.status.Set(Status{
		State:   StateStarting,
		Channel: channel,
	})

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

	streamID := fmt.Sprintf("Tuner(%p)", t)
	vt, at, err := createTrackPair(streamID)
	if err != nil {
		return
	}

	t.pipeline.SetSink(gst.SinkTypeVideo, createTrackSink(vt, videoClockRate))
	t.pipeline.SetSink(gst.SinkTypeAudio, createTrackSink(at, audioClockRate))

	log.Printf("Tuner(%p): Starting pipeline", t)
	err = t.pipeline.Start()
	if err != nil {
		return
	}
	log.Printf("Tuner(%p): Started pipeline", t)

	t.status.Set(Status{
		State:   StatePlaying,
		Channel: channel,
	})
	t.tracks.Set(Tracks{
		Video: vt,
		Audio: at,
	})
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

func createTrackSink(track *webrtc.Track, clockRate int) gst.Sink {
	return gst.Sink(func(b []byte, d time.Duration) {
		track.WriteSample(media.Sample{
			Data:    b,
			Samples: media.NSamples(d, clockRate),
		})
	})
}
