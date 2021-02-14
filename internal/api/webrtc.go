package api

import (
	"encoding/json"
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
	var (
		me  webrtc.MediaEngine
		err error
	)

	err = me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: tuner.VideoCodecCapability,
		PayloadType:        tuner.VideoPayloadType,
	}, webrtc.RTPCodecTypeVideo)
	if err != nil {
		panic(err)
	}

	err = me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: tuner.AudioCodecCapability,
		PayloadType:        tuner.AudioPayloadType,
	}, webrtc.RTPCodecTypeAudio)
	if err != nil {
		panic(err)
	}

	webrtcAPI = webrtc.NewAPI(webrtc.WithMediaEngine(&me))
}

type webrtcHandler struct {
	tuner *tuner.Tuner

	conn        *websocket.Conn
	pc          *webrtc.PeerConnection
	watch       *watch.Watch
	shutdownErr chan error
	wg          sync.WaitGroup
}

func (h *Handler) handleSocketWebRTCPeer(w http.ResponseWriter, r *http.Request) {
	wh := &webrtcHandler{
		tuner:       h.tuner,
		shutdownErr: make(chan error, 1),
	}

	wh.ServeHTTP(w, r)
}

func (wh *webrtcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	wh.logf("Starting new connection")
	defer func() {
		wh.waitForCleanup()
		wh.logf("Finished with error: %v", err)
	}()

	wh.conn, err = websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer wh.conn.Close()

	wh.pc, err = webrtcAPI.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return
	}
	defer wh.pc.Close()

	// TODO: Don't try to negotiate with the client before we've defined video and
	// audio transceivers. This is a hack to reproduce Pion WebRTC v2 behavior.
	_, err = wh.pc.CreateDataChannel("unused", nil)
	if err != nil {
		return
	}

	wh.wg.Add(1)
	go func() {
		defer wh.wg.Done()
		wh.handleClientSessionAnswers()
	}()

	wh.watch = wh.tuner.WatchTracks(wh.handleTrackUpdate)
	defer wh.watch.Cancel()

	err = <-wh.shutdownErr
	return
}

func (wh *webrtcHandler) handleClientSessionAnswers() {
	for {
		_, r, err := wh.conn.NextReader()
		if err != nil {
			wh.shutdown(err)
			return
		}

		var msg struct{ SDP webrtc.SessionDescription }
		if err := json.NewDecoder(r).Decode(&msg); err != nil {
			wh.shutdown(err)
			return
		}

		if err := wh.pc.SetRemoteDescription(msg.SDP); err != nil {
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
	sdp, err := wh.pc.CreateOffer(nil)
	if err != nil {
		return err
	}

	// TODO: Trickle ICE
	gatherComplete := webrtc.GatheringCompletePromise(wh.pc)

	if err := wh.pc.SetLocalDescription(sdp); err != nil {
		return err
	}

	<-gatherComplete
	msg := struct{ SDP webrtc.SessionDescription }{*wh.pc.LocalDescription()}
	return wh.conn.WriteJSON(msg)
}

func (wh *webrtcHandler) removeTracks() error {
	for _, sender := range wh.pc.GetSenders() {
		if err := wh.pc.RemoveTrack(sender); err != nil {
			return err
		}
	}
	return nil
}

func (wh *webrtcHandler) addTracks(ts tuner.Tracks) error {
	if len(wh.pc.GetTransceivers()) == 0 {
		return wh.addTracksWithNewTransceivers(ts)
	}

	return wh.addTracksWithExistingTransceivers(ts)
}

func (wh *webrtcHandler) addTracksWithNewTransceivers(ts tuner.Tracks) error {
	init := webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}

	if _, err := wh.pc.AddTransceiverFromTrack(ts.Video, init); err != nil {
		return err
	}
	if _, err := wh.pc.AddTransceiverFromTrack(ts.Audio, init); err != nil {
		return err
	}
	return nil
}

func (wh *webrtcHandler) addTracksWithExistingTransceivers(ts tuner.Tracks) error {
	if _, err := wh.pc.AddTrack(ts.Video); err != nil {
		return err
	}
	if _, err := wh.pc.AddTrack(ts.Audio); err != nil {
		return err
	}
	return nil
}

func (wh *webrtcHandler) shutdown(err error) {
	select {
	case wh.shutdownErr <- err:
	default:
	}
}

func (wh *webrtcHandler) waitForCleanup() {
	if wh.watch != nil {
		wh.watch.Wait()
	}
	wh.wg.Wait()
}

func (wh *webrtcHandler) logf(format string, v ...interface{}) {
	joinFmt := "WebRTCHandler(%p): " + format

	joinArgs := make([]interface{}, len(v)+1)
	joinArgs[0] = wh
	copy(joinArgs[1:], v)

	log.Printf(joinFmt, joinArgs...)
}
