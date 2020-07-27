package client

import "github.com/pion/webrtc/v2"

type messageKind string

const (
	// Server-sent messages
	messageKindRTCOffer    messageKind = "RTCOffer"
	messageKindChannelList messageKind = "ChannelList"
	messageKindTunerStatus messageKind = "TunerStatus"

	// Client-sent messages
	messageKindRTCAnswer     messageKind = "RTCAnswer"
	messageKindChangeChannel messageKind = "ChangeChannel"
	messageKindTurnOff       messageKind = "TurnOff"
)

type message struct {
	Kind messageKind

	// RTCOffer, RTCAnswer
	SDP *webrtc.SessionDescription `json:",omitempty"`
	// ChannelList
	ChannelNames []string `json:",omitempty"`
	// TunerStatus
	TunerStatus *tunerStatus `json:",omitempty"`
	// ChangeChannel
	ChannelName string `json:",omitempty"`
}

type tunerStatus struct {
	ChannelName string `json:",omitempty"`
	Error       error  `json:",omitempty"`
}
