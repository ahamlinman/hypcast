package oldclient

import "github.com/pion/webrtc/v2"

type messageKind string

const (
	// Server-sent messages
	messageKindRTCOffer    messageKind = "RTCOffer"
	messageKindTunerStatus messageKind = "TunerStatus"

	// Client-sent messages
	messageKindRTCAnswer messageKind = "RTCAnswer"
)

type message struct {
	Kind messageKind

	// RTCOffer, RTCAnswer
	SDP *webrtc.SessionDescription `json:",omitempty"`
	// TunerStatus
	TunerStatus *tunerStatus `json:",omitempty"`
}

type tunerStatus struct {
	State       string `json:",omitempty"`
	ChannelName string `json:",omitempty"`
	Error       error  `json:",omitempty"`
}
