package tuner

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sync"
	"text/template"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"

	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/gst"
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
	State       State
	ChannelName string
	Error       error
}

// Tracks represents the current set of video and audio tracks for use by WebRTC
// clients.
type Tracks struct {
	Video webrtc.TrackLocal
	Audio webrtc.TrackLocal
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

// Stop ends any active stream and releases the DVB device associated with this
// tuner.
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

// Tune attempts to start a stream for the named channel.
func (t *Tuner) Tune(channelName string) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	channel, ok := t.channelMap[channelName]
	if !ok {
		return ErrChannelNotFound
	}

	t.status.Set(Status{
		State:       StateStarting,
		ChannelName: channel.Name,
	})

	defer func() {
		if err != nil {
			t.destroyAnyRunningPipeline()
			t.status.Set(Status{Error: err})
		}
	}()

	t.destroyAnyRunningPipeline()

	t.pipeline, err = newPipeline(channel)
	if err != nil {
		return err
	}

	vt, at, err := t.createTrackPair()
	if err != nil {
		return err
	}

	t.pipeline.SetSink(sinkNameVideo, createTrackSink(vt))
	t.pipeline.SetSink(sinkNameAudio, createTrackSink(at))

	log.Printf("Tuner(%p): Starting pipeline", t)
	err = t.pipeline.Start()
	if err != nil {
		return err
	}
	log.Printf("Tuner(%p): Started pipeline", t)

	t.status.Set(Status{
		State:       StatePlaying,
		ChannelName: channelName,
	})
	t.tracks.Set(Tracks{
		Video: vt,
		Audio: at,
	})
	return nil
}

func newPipeline(channel atsc.Channel) (*gst.Pipeline, error) {
	description, err := createPipelineDescription(channel)
	if err != nil {
		return nil, err
	}
	return gst.NewPipeline(description)
}

func createPipelineDescription(channel atsc.Channel) (string, error) {
	var buf bytes.Buffer

	err := pipelineDescriptionTemplate.Execute(&buf, struct {
		Modulation string
		Frequency  uint
		PID        uint
	}{
		Modulation: pipelineModulations[channel.Modulation],
		Frequency:  channel.Frequency,
		PID:        channel.ProgramID,
	})
	if err != nil {
		return "", fmt.Errorf("building pipeline template: %w", err)
	}

	return buf.String(), nil
}

var pipelineModulations = map[atsc.Modulation]string{
	atsc.Modulation8VSB:   "8vsb",
	atsc.ModulationQAM64:  "qam-64",
	atsc.ModulationQAM256: "qam-256",
}

// These match up with the names of appsink elements in the pipeline
// description below.
const (
	sinkNameRaw   = "raw"
	sinkNameVideo = "video"
	sinkNameAudio = "audio"
)

// TODO:
// https://gitlab.freedesktop.org/gstreamer/gstreamer/-/issues/358#note_118032
// Without drop-allocation the pipeline stalls. I still don't *really*
// understand why.
var pipelineDescriptionTemplate = template.Must(template.New("").Parse(`
	dvbsrc delsys=atsc modulation={{.Modulation}} frequency={{.Frequency}}
	! tee name=dvbtee
	! identity drop-allocation=true
	! queue leaky=downstream max-size-time=1000000000 max-size-buffers=0 max-size-bytes=0
	! appsink name=raw max-buffers=32 drop=true

	dvbtee.
	! queue leaky=downstream max-size-time=0 max-size-buffers=0 max-size-bytes=0
	! tsdemux name=demux program-number={{.PID}}

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! videoconvert
	! deinterlace
	! x264enc bitrate=8192 tune=zerolatency speed-preset=ultrafast
	! video/x-h264,profile=constrained-baseline,stream-format=byte-stream
	! appsink name=video max-buffers=32 drop=true

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! audioconvert
	! audioresample
	! audio/x-raw,rate=48000,channels=2
	! opusenc bitrate=128000
	! appsink name=audio max-buffers=32 drop=true
`))

func (t *Tuner) destroyAnyRunningPipeline() error {
	if t.pipeline == nil {
		return nil
	}

	err := t.pipeline.Close()
	t.pipeline = nil
	return err
}

// fmtp is described by https://tools.ietf.org/html/rfc6184.
//
// profile-level-id in particular is described in section 8.1 of the RFC. The
// first 2 octets together indicate the Constrained Baseline profile (42h to
// specify the Baseline profile, e0h to specify constraint set 1). The third
// octet (28h = 40) specifies level 4.0 (the level number times 10), the lowest
// to support 1920x1080 video per
// https://en.wikipedia.org/wiki/Advanced_Video_Coding#Levels.
//
// This needs to match up with the pipeline definition in the gst package.
const videoCodecFMTP = "profile-level-id=42e028;level-asymmetry-allowed=1;packetization-mode=1"

var (
	// VideoCodecCapability represents the RTP codec settings for the video signal
	// produced by the tuner.
	VideoCodecCapability = webrtc.RTPCodecCapability{
		MimeType:    webrtc.MimeTypeH264,
		ClockRate:   90_000,
		SDPFmtpLine: videoCodecFMTP,
	}

	// AudioCodecCapability represents the RTP codec settings for the audio signal
	// produced by the tuner.
	AudioCodecCapability = webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48_000,
		Channels:  2,
	}
)

func (t *Tuner) createTrackPair() (video, audio *webrtc.TrackLocalStaticSample, err error) {
	streamID := fmt.Sprintf("Tuner(%p)", t)

	video, err = webrtc.NewTrackLocalStaticSample(VideoCodecCapability, streamID, streamID)
	if err != nil {
		return
	}

	audio, err = webrtc.NewTrackLocalStaticSample(AudioCodecCapability, streamID, streamID)
	return
}

func createTrackSink(track *webrtc.TrackLocalStaticSample) gst.SinkFunc {
	return gst.SinkFunc(func(data []byte, duration time.Duration) {
		track.WriteSample(media.Sample{
			Data:     data,
			Duration: duration,
		})
	})
}
