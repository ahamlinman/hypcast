package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

var upgrader = websocket.Upgrader{
	// FIXME: Unsafe; for testing purposes only
	CheckOrigin: func(_ *http.Request) bool { return true },
}

const (
	videoClockRate = 90_000
	audioClockRate = 48_000
)

type socketHandler struct {
	videoTrack *webrtc.Track
	audioTrack *webrtc.Track
}

func newSocketHandler() (*socketHandler, error) {
	videoTrack, err := webrtc.NewTrack(
		webrtc.DefaultPayloadTypeVP8, 5000, "hyp", "hyp",
		webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, videoClockRate),
	)
	if err != nil {
		return nil, err
	}

	audioTrack, err := webrtc.NewTrack(
		webrtc.DefaultPayloadTypeOpus, 5001, "hyp", "hyp",
		webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, audioClockRate),
	)
	if err != nil {
		return nil, err
	}

	return &socketHandler{
		videoTrack: videoTrack,
		audioTrack: audioTrack,
	}, nil
}

func (h *socketHandler) HandleAudioData(buffer []byte, d time.Duration) {
	h.audioTrack.WriteSample(media.Sample{
		Data:    buffer,
		Samples: media.NSamples(d, audioClockRate),
	})
}

func (h *socketHandler) HandleVideoData(buffer []byte, d time.Duration) {
	h.videoTrack.WriteSample(media.Sample{
		Data:    buffer,
		Samples: media.NSamples(d, videoClockRate),
	})
}

func (h *socketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request(%p): Received", r)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Request(%p): Failed to upgrade connection: %v", r, err)
		return
	}
	log.Printf("Request(%p): Upgraded connection", r)
	defer ws.Close()

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Printf("Request(%p): Failed to create PeerConnection: %v", r, err)
		return
	}
	log.Printf("Request(%p): Created PeerConnection", r)
	defer pc.Close()

	if _, err = pc.AddTrack(h.videoTrack); err != nil {
		log.Printf("Request(%p): Failed to add video track: %v", r, err)
		return
	}

	if _, err = pc.AddTrack(h.audioTrack); err != nil {
		log.Printf("Request(%p): Failed to add audio track: %v", r, err)
		return
	}

	log.Printf("Request(%p): Added tracks", r)

	sdp, err := pc.CreateOffer(nil)
	if err != nil {
		log.Printf("Request(%p): Failed to create local session description: %v", r, err)
		return
	}

	if err := pc.SetLocalDescription(sdp); err != nil {
		log.Printf("Request(%p): Failed to set local session description: %v", r, err)
		return
	}
	log.Printf("Request(%p): Set local session description", r)

	msg := message{
		Kind:        serverOfferMessageKind,
		ServerOffer: &sdp,
	}
	if err := ws.WriteJSON(msg); err != nil {
		log.Printf("Request(%p): Failed to send offer to client: %v", r, err)
		return
	}

	for {
		_, msgData, err := ws.ReadMessage()
		if err != nil {
			if cerr, ok := err.(*websocket.CloseError); ok {
				log.Printf("Request(%p): Client disconnected [%s]", r, cerr.Error())
			} else {
				log.Printf("Request(%p): Unexpected error reading client message: %v", r, err)
			}

			return
		}

		var msg message
		if err := json.Unmarshal(msgData, &msg); err != nil {
			log.Printf("Request(%p): Received invalid message: %v", r, err)
			return
		}

		switch msg.Kind {
		case clientAnswerMessageKind:
			log.Printf("Request(%p): Received session description from client", r)
			pc.SetRemoteDescription(*msg.ClientAnswer)

		default:
			log.Printf("Request(%p): Ignoring unknown message kind %q", r, msg.Kind)
		}
	}
}

type messageKind string

const (
	serverOfferMessageKind  messageKind = "ServerOffer"
	clientAnswerMessageKind             = "ClientAnswer"
)

type message struct {
	Kind         messageKind
	ServerOffer  *webrtc.SessionDescription `json:",omitempty"`
	ClientAnswer *webrtc.SessionDescription `json:",omitempty"`
}
