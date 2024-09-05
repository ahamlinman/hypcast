package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
	"github.com/ahamlinman/hypcast/internal/watch"
)

var webrtcAPI *webrtc.API

func init() {
	// https://tools.ietf.org/html/rfc3551#section-3
	//
	// "This profile reserves payload type numbers in the range 96-127 exclusively
	// for dynamic assignment."
	const (
		videoPayloadType = 96 + iota
		audioPayloadType
	)

	var me webrtc.MediaEngine
	err := errors.Join(
		me.RegisterCodec(
			webrtc.RTPCodecParameters{PayloadType: videoPayloadType, RTPCodecCapability: tuner.VideoCodecCapability},
			webrtc.RTPCodecTypeVideo),
		me.RegisterCodec(
			webrtc.RTPCodecParameters{PayloadType: audioPayloadType, RTPCodecCapability: tuner.AudioCodecCapability},
			webrtc.RTPCodecTypeAudio),
	)
	if err != nil {
		panic(err)
	}

	webrtcAPI = webrtc.NewAPI(webrtc.WithMediaEngine(&me))
}

type webrtcHandler struct {
	tuner     *tuner.Tuner
	socket    *websocket.Conn
	rtcPeer   *webrtc.PeerConnection
	watch     watch.Watch
	ctx       context.Context
	shutdown  context.CancelCauseFunc
	waitGroup sync.WaitGroup
}

func (h *Handler) handleSocketWebRTCPeer(w http.ResponseWriter, r *http.Request) {
	ctx, shutdown := context.WithCancelCause(r.Context())
	wh := &webrtcHandler{
		tuner:    h.tuner,
		ctx:      ctx,
		shutdown: shutdown,
	}
	wh.ServeHTTP(w, r)
}

func (wh *webrtcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wh.logf("Starting new connection")
	defer func() {
		wh.waitForCleanup()
		wh.logf("Connection done: %v", context.Cause(wh.ctx))
	}()

	var err error
	wh.socket, err = websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer wh.socket.Close()

	wh.rtcPeer, err = webrtcAPI.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return
	}
	defer wh.rtcPeer.Close()

	wh.waitGroup.Add(1)
	go func() {
		defer wh.waitGroup.Done()
		wh.handleClientSessionAnswers()
	}()

	wh.watch = wh.tuner.WatchTracks(wh.handleTrackUpdate)
	defer wh.watch.Cancel()

	<-wh.ctx.Done()
}

func (wh *webrtcHandler) handleClientSessionAnswers() {
	for {
		_, r, err := wh.socket.NextReader()
		if err != nil {
			wh.shutdown(err)
			return
		}

		var msg struct{ SDP webrtc.SessionDescription }
		if err := json.NewDecoder(r).Decode(&msg); err != nil {
			wh.shutdown(err)
			return
		}

		if err := wh.rtcPeer.SetRemoteDescription(msg.SDP); err != nil {
			wh.shutdown(err)
			return
		}
	}
}

func (wh *webrtcHandler) handleTrackUpdate(ts tuner.Tracks) {
	wh.logf("Received tracks: %v", ts)
	if err := wh.replaceTracks(ts); err != nil {
		wh.shutdown(err)
		return
	}
	if err := wh.renegotiateSession(); err != nil {
		wh.shutdown(err)
		return
	}
}

func (wh *webrtcHandler) replaceTracks(ts tuner.Tracks) error {
	if err := wh.removeTracks(); err != nil {
		return err
	}
	if ts == (tuner.Tracks{}) {
		return nil
	}
	return wh.addTracks(ts)
}

func (wh *webrtcHandler) renegotiateSession() error {
	if !wh.hasTransceivers() {
		// Skip negotiation until we've had a chance to properly define video and
		// audio transceivers based on Tuner tracks.
		return nil
	}

	sdp, err := wh.rtcPeer.CreateOffer(nil)
	if err != nil {
		return err
	}

	// TODO: Should probably implement trickle ICE, but since Hypcast doesn't
	// implement STUN support it's not like ICE gathering takes much time.
	gatherComplete := webrtc.GatheringCompletePromise(wh.rtcPeer)

	if err := wh.rtcPeer.SetLocalDescription(sdp); err != nil {
		return err
	}

	<-gatherComplete
	msg := struct{ SDP webrtc.SessionDescription }{*wh.rtcPeer.LocalDescription()}
	return wh.socket.WriteJSON(msg)
}

func (wh *webrtcHandler) removeTracks() error {
	for _, sender := range wh.rtcPeer.GetSenders() {
		if err := wh.rtcPeer.RemoveTrack(sender); err != nil {
			return err
		}
	}
	return nil
}

func (wh *webrtcHandler) addTracks(ts tuner.Tracks) error {
	if wh.hasTransceivers() {
		return wh.addTracksWithExistingTransceivers(ts)
	}
	return wh.addTracksWithNewTransceivers(ts)
}

func (wh *webrtcHandler) addTracksWithExistingTransceivers(ts tuner.Tracks) error {
	if _, err := wh.rtcPeer.AddTrack(ts.Video); err != nil {
		return err
	}
	if _, err := wh.rtcPeer.AddTrack(ts.Audio); err != nil {
		return err
	}
	return nil
}

func (wh *webrtcHandler) addTracksWithNewTransceivers(ts tuner.Tracks) error {
	init := webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}
	if _, err := wh.rtcPeer.AddTransceiverFromTrack(ts.Video, init); err != nil {
		return err
	}
	if _, err := wh.rtcPeer.AddTransceiverFromTrack(ts.Audio, init); err != nil {
		return err
	}
	return nil
}

func (wh *webrtcHandler) hasTransceivers() bool {
	return len(wh.rtcPeer.GetTransceivers()) > 0
}

func (wh *webrtcHandler) waitForCleanup() {
	if wh.watch != nil {
		wh.watch.Wait()
	}
	wh.waitGroup.Wait()
}

func (wh *webrtcHandler) logf(format string, v ...any) {
	joinFmt := "WebRTCHandler(%p): " + format
	joinArgs := make([]any, len(v)+1)
	joinArgs[0] = wh
	copy(joinArgs[1:], v)
	log.Printf(joinFmt, joinArgs...)
}
