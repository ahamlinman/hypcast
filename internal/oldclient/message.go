package oldclient

import "github.com/pion/webrtc/v2"

type messageKind string

const (
	messageKindRTCOffer  messageKind = "RTCOffer"
	messageKindRTCAnswer messageKind = "RTCAnswer"
)

type message struct {
	Kind messageKind
	SDP  *webrtc.SessionDescription `json:",omitempty"`
}
