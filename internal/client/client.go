package client

import (
	"encoding/json"
	"fmt"
	"log"
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

		log.Printf("Client(%p): Starting", client)
		err = client.start()
		log.Printf("Client(%p): Finished with error: %v", client, err)
	})
}

type client struct {
	tuner *tuner.Tuner

	ws *websocket.Conn
	pc *webrtc.PeerConnection

	videoTrack *webrtc.Track
	audioTrack *webrtc.Track

	receiverDone       chan error
	rtcOfferAvailable  chan webrtc.SessionDescription
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

	return c.runSender()
}

func (c *client) init() error {
	c.receiverDone = make(chan error, 1)

	var err error
	c.pc, err = webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}

	c.rtcOfferAvailable = make(chan webrtc.SessionDescription, 1)

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
		_, rawMsg, err := c.ws.ReadMessage()
		if err != nil {
			return err
		}

		var msg message
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			return err
		}

		switch msg.Kind {
		case messageKindRTCAnswer:
			if err := c.pc.SetRemoteDescription(*msg.SDP); err != nil {
				return err
			}

		case messageKindChangeChannel:
			if err := c.tuner.Tune(msg.ChannelName); err != nil {
				return err
			}

		case messageKindTurnOff:
			if err := c.tuner.Stop(); err != nil {
				return err
			}

		default:
			return fmt.Errorf("received unknown message kind %q", msg.Kind)
		}
	}
}

func (c *client) runSender() error {
	for {
		select {
		case err := <-c.receiverDone:
			return err

		case <-c.tunerSyncRequested:
			if err := c.syncTunerStatus(); err != nil {
				return err
			}
		}
	}
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

	sdp, err := c.pc.CreateOffer(nil)
	if err != nil {
		return err
	}
	if err := c.writeOfferMessage(sdp); err != nil {
		return err
	}

	return c.writeTunerStatusMessage(s)
}

func (c *client) removeExistingTracks() error {
	for _, sender := range c.pc.GetSenders() {
		if err := c.pc.RemoveTrack(sender); err != nil {
			return err
		}
	}

	c.videoTrack = nil
	c.audioTrack = nil
	return nil
}

func (c *client) addNewTracks(video, audio *webrtc.Track) error {
	if _, err := c.pc.AddTrack(video); err != nil {
		return err
	}

	if _, err := c.pc.AddTrack(audio); err != nil {
		return err
	}

	c.videoTrack = video
	c.audioTrack = audio
	return nil
}

func (c *client) writeOfferMessage(sdp webrtc.SessionDescription) error {
	return c.ws.WriteJSON(message{
		Kind: messageKindRTCOffer,
		SDP:  &sdp,
	})
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
