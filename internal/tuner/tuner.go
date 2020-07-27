package tuner

import (
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"

	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/gst"
)

// Client receives notifications when the status of a Tuner is updated.
type Client interface {
	CheckTunerStatus()
}

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
	channels   []atsc.Channel
	channelMap map[string]atsc.Channel

	mu       sync.Mutex
	clients  map[Client]struct{}
	pipeline *gst.Pipeline
	status   Status
}

// NewTuner creates a new Tuner that can tune to any of the provided channels.
func NewTuner(channels []atsc.Channel) *Tuner {
	return &Tuner{
		clients:    make(map[Client]struct{}),
		channels:   channels,
		channelMap: makeChannelMap(channels),
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
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.status
}

// AddClient registers a Client to be notified when the state of the Tuner has
// been updated.
func (t *Tuner) AddClient(client Client) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.clients[client] = struct{}{}
}

// RemoveClient removes this Client from further state update notifications.
func (t *Tuner) RemoveClient(client Client) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.clients, client)
}

// Stop closes any active pipeline for this Tuner, releasing the DVB device, and
// notifies all clients of the change.
func (t *Tuner) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.notifyClients()

	err := t.stop()
	t.status.Error = err
	return err
}

// Tune closes any active pipeline for this Tuner, starts a new pipeline to
// stream the channel with the provided name, and notifies all clients of the
// change.
func (t *Tuner) Tune(channelName string) (err error) {
	channel, ok := t.channelMap[channelName]
	if !ok {
		return fmt.Errorf("channel %q not available in this tuner", channelName)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	defer t.notifyClients()

	t.stop()

	defer func() {
		if err != nil {
			t.stop()
			t.status.Error = err
		}
	}()

	t.pipeline, err = gst.NewPipeline(channel)
	if err != nil {
		return
	}

	streamID := fmt.Sprintf("Track(%p)", t)
	t.status.VideoTrack, t.status.AudioTrack, err = createTrackPair(streamID)
	if err != nil {
		return
	}

	t.pipeline.SetSink(gst.SinkTypeVideo, sinkTrack(t.status.VideoTrack, videoClockRate))
	t.pipeline.SetSink(gst.SinkTypeAudio, sinkTrack(t.status.AudioTrack, audioClockRate))

	t.status.Active = true
	t.status.Channel = channel

	err = t.pipeline.Start()
	return
}

func (t *Tuner) notifyClients() {
	for c := range t.clients {
		c.CheckTunerStatus()
	}
}

func (t *Tuner) stop() error {
	defer func() {
		t.status = Status{}
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

func createTrackPair(streamID string) (video *webrtc.Track, audio *webrtc.Track, err error) {
	ssrcBase := generateSSRCBase()

	video, err = webrtc.NewTrack(
		webrtc.DefaultPayloadTypeVP8, ssrcBase, streamID, streamID,
		webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, videoClockRate),
	)
	if err != nil {
		return
	}

	audio, err = webrtc.NewTrack(
		webrtc.DefaultPayloadTypeOpus, ssrcBase+1, streamID, streamID,
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

var (
	ssrcMu   sync.Mutex
	ssrcRand *rand.Rand
)

func init() {
	max := big.NewInt(math.MaxInt64)
	seed, err := cryptorand.Int(cryptorand.Reader, max)
	if err != nil {
		panic(fmt.Errorf("failed to init random ssrc generator: %w", err))
	}

	source := rand.NewSource(seed.Int64())
	ssrcRand = rand.New(source)
}

func generateSSRCBase() uint32 {
	ssrcMu.Lock()
	defer ssrcMu.Unlock()

	return ssrcRand.Uint32() &^ 1
}
