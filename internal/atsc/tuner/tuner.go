package tuner

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pion/randutil"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"

	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/tuner/gst"
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
	Channel atsc.Channel // TODO: Only expose channel name
	Error   error
}

// Tracks represents the current set of video and audio tracks for use by WebRTC
// clients.
type Tracks struct {
	Video *webrtc.Track
	Audio *webrtc.Track
}

// Tuner represents an ATSC tuner whose video and audio signals are encoded for
// use by WebRTC clients, and whose consumers are notified of ongoing state
// changes.
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

// ChannelNames returns the names of channels that are known to this tuner and
// may be passed to Tune.
func (t *Tuner) ChannelNames() []string {
	channelNames := make([]string, len(t.channels))
	for i, ch := range t.channels {
		channelNames[i] = ch.Name
	}
	return channelNames
}

// WatchStatus sets up a handler function to continuously receive the status of
// the tuner as it is updated. See the watch package documentation for details.
func (t *Tuner) WatchStatus(handler func(Status)) *watch.Watch {
	return t.status.Watch(func(x interface{}) {
		handler(x.(Status))
	})
}

// WatchTracks sets up a handler function to continuously receive the tuner's
// WebRTC tracks as they are updated. See the watch package documentation for
// details.
func (t *Tuner) WatchTracks(handler func(Tracks)) *watch.Watch {
	return t.tracks.Watch(func(x interface{}) {
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

	t.pipeline.SetSink(gst.SinkTypeVideo, createTrackSink(vt))
	t.pipeline.SetSink(gst.SinkTypeAudio, createTrackSink(at))

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

	// All of this is described by https://tools.ietf.org/html/rfc6184.
	//
	// profile-level-id in particular is described in section 8.1 of the RFC. The
	// first 2 octets together indicate the Constrained Baseline profile (42h to
	// specify the Baseline profile, e0h to specify constraint set 1). The third
	// octet (28h = 40) specifies level 4.0 (the level number times 10), the
	// lowest to support 1920x1080 video per
	// https://en.wikipedia.org/wiki/Advanced_Video_Coding#Levels.
	//
	// This needs to match up with the pipeline definition in the gst package.
	videoCodecFMTP = "profile-level-id=42e028;level-asymmetry-allowed=1;packetization-mode=1"
)

var (
	// VideoCodec represents the RTP codec settings for the video signal produced
	// by the tuner.
	VideoCodec = webrtc.NewRTPCodec(
		webrtc.RTPCodecTypeVideo, webrtc.H264, videoClockRate, 0, videoCodecFMTP,
		webrtc.DefaultPayloadTypeH264, &codecs.H264Payloader{},
	)
	// AudioCodec represents the RTP codec settings for the audio signal produced
	// by the tuner.
	AudioCodec = webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, audioClockRate)

	ssrcGenerator = randutil.NewMathRandomGenerator()
)

func createTrackPair(streamID string) (video *webrtc.Track, audio *webrtc.Track, err error) {
	video, err = webrtc.NewTrack(
		webrtc.DefaultPayloadTypeH264, ssrcGenerator.Uint32(), streamID, streamID, VideoCodec,
	)
	if err != nil {
		return
	}

	audio, err = webrtc.NewTrack(
		webrtc.DefaultPayloadTypeOpus, ssrcGenerator.Uint32(), streamID, streamID, AudioCodec,
	)
	return
}

func createTrackSink(track *webrtc.Track) gst.Sink {
	clockRate := int(track.Codec().ClockRate)
	return gst.Sink(func(b []byte, d time.Duration) {
		track.WriteSample(media.Sample{
			Data:    b,
			Samples: media.NSamples(d, clockRate),
		})
	})
}
