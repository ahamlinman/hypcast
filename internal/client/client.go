package client

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"

	"github.com/ahamlinman/hypcast/internal/tuner"
)

var upgrader = websocket.Upgrader{
	// TODO: For testing purposes only.
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Handler returns a http.Handler that spawns a new tuner client for each new
// WebSocket connection.
func Handler(tuner *tuner.Tuner) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &client{
			tuner: tuner,
			ws:    ws,
		}
		client.start()
	})
}

type client struct {
	tuner *tuner.Tuner

	ws *websocket.Conn
	pc *webrtc.PeerConnection

	videoTrack  *webrtc.Track
	videoSender *webrtc.RTPSender
	audioTrack  *webrtc.Track
	audioSender *webrtc.RTPSender

	receiverDone       chan error
	rtcOfferAvailable  chan webrtc.SessionDescription
	rtcAnswerAvailable chan webrtc.SessionDescription
	tunerSyncRequested chan struct{}
}

func (c *client) start() error {
	defer func() {
		c.tuner.RemoveClient(c)
		if c.pc != nil {
			c.pc.Close()
		}
		c.ws.Close()
	}()

	if err := c.init(); err != nil {
		return err
	}

	if err := c.writeChannelListMessage(); err != nil {
		return err
	}

	go func() {
		defer close(c.receiverDone)
		if err := c.runReceiver(); err != nil {
			c.receiverDone <- err
		}
	}()

	return c.runHandler()
}

func (c *client) init() error {
	c.receiverDone = make(chan error, 1)

	var err error
	c.pc, err = webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}

	sdp, err := c.pc.CreateOffer(nil)
	if err != nil {
		return err
	}
	c.pc.SetLocalDescription(sdp)
	c.rtcOfferAvailable = make(chan webrtc.SessionDescription, 1)
	c.rtcOfferAvailable <- sdp

	c.rtcAnswerAvailable = make(chan webrtc.SessionDescription, 1)

	c.tunerSyncRequested = make(chan struct{}, 1)
	c.tunerSyncRequested <- struct{}{}

	c.tuner.AddClient(c)
	return nil
}

func (c *client) writeChannelListMessage() error {
	var channelNames []string
	for _, ch := range c.tuner.Channels() {
		channelNames = append(channelNames, ch.Name)
	}

	return c.ws.WriteJSON(message{
		Kind:         messageKindChannelList,
		ChannelNames: channelNames,
	})
}

func (c *client) runReceiver() error {
	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			return err
		}
	}
}

func (c *client) runHandler() error {
	for {
		select {
		case err := <-c.receiverDone:
			return err

		case sdp := <-c.rtcOfferAvailable:
			if err := c.writeOfferMessage(sdp); err != nil {
				return err
			}

		case sdp := <-c.rtcAnswerAvailable:
			if err := c.pc.SetRemoteDescription(sdp); err != nil {
				return err
			}

		case <-c.tunerSyncRequested:
			if err := c.syncTunerStatus(); err != nil {
				return err
			}
		}
	}
}

func (c *client) writeOfferMessage(sdp webrtc.SessionDescription) error {
	return c.ws.WriteJSON(message{
		Kind: messageKindRTCOffer,
		SDP:  &sdp,
	})
}

func (c *client) syncTunerStatus() error {
	s := c.tuner.Status()

	tracksChanged := s.VideoTrack != c.videoTrack || s.AudioTrack != c.audioTrack

	if !s.Active || tracksChanged {
		if err := c.removeExistingTracks(); err != nil {
			return err
		}
	}

	if s.Active && tracksChanged {
		if err := c.addNewTracks(s.VideoTrack, s.AudioTrack); err != nil {
			return err
		}
	}

	return c.writeTunerStatusMessage(s)
}

func (c *client) removeExistingTracks() error {
	if c.videoSender != nil {
		if err := c.pc.RemoveTrack(c.videoSender); err != nil {
			return err
		}
		c.videoSender = nil
		c.videoTrack = nil
	}

	if c.audioSender != nil {
		if err := c.pc.RemoveTrack(c.audioSender); err != nil {
			return err
		}
		c.audioSender = nil
		c.audioTrack = nil
	}

	return nil
}

func (c *client) addNewTracks(video, audio *webrtc.Track) error {
	var err error

	c.videoTrack = video
	c.videoSender, err = c.pc.AddTrack(video)
	if err != nil {
		return err
	}

	c.audioTrack = audio
	c.audioSender, err = c.pc.AddTrack(audio)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) writeTunerStatusMessage(s tuner.Status) error {
	return c.ws.WriteJSON(message{
		Kind: messageKindTunerStatus,
		TunerStatus: &tunerStatus{
			ChannelName: s.Channel.Name,
			Error:       s.Error,
		},
	})
}

func (c *client) CheckTunerStatus() {
	select {
	case c.tunerSyncRequested <- struct{}{}:
	default:
	}
}
