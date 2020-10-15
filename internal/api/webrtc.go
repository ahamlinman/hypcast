package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
)

func (h *Handler) handleSocketWebRTCPeer(w http.ResponseWriter, r *http.Request) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	wh := &webrtcHandler{
		conn:  conn,
		tuner: h.tuner,
	}
	wh.run()
}

type webrtcHandler struct {
	conn       *websocket.Conn
	tuner      *tuner.Tuner
	pc         *webrtc.PeerConnection
	err        chan error
	clientDone chan struct{}
}

func (wh *webrtcHandler) run() (err error) {
	wh.logf("Starting new connection")
	defer func() { wh.logf("Finished with error: %v", err) }()

	wh.err = make(chan error, 1)
	wh.clientDone = make(chan struct{})

	wh.pc, err = webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return
	}
	defer wh.pc.Close()

	go wh.handleClientSessionAnswers()
	defer func() { <-wh.clientDone }()

	s := wh.tuner.SubscribeTracks(wh.handleTrackUpdate)
	defer func() {
		s.Cancel()
		s.Wait()
	}()

	return <-wh.err
}

func (wh *webrtcHandler) handleClientSessionAnswers() {
	defer close(wh.clientDone)
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

	if err := wh.pc.SetLocalDescription(sdp); err != nil {
		return err
	}

	msg := struct{ SDP webrtc.SessionDescription }{sdp}
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
	init := webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}

	if _, err := wh.pc.AddTransceiverFromTrack(ts.VideoTrack, init); err != nil {
		return err
	}
	if _, err := wh.pc.AddTransceiverFromTrack(ts.AudioTrack, init); err != nil {
		return err
	}
	return nil
}

func (wh *webrtcHandler) addTracksWithExistingTransceivers(ts tuner.Tracks) error {
	if _, err := wh.pc.AddTrack(ts.VideoTrack); err != nil {
		return err
	}
	if _, err := wh.pc.AddTrack(ts.AudioTrack); err != nil {
		return err
	}
	return nil
}

func (wh *webrtcHandler) shutdown(err error) {
	select {
	case wh.err <- err:
	default:
	}
}

func (wh *webrtcHandler) logf(format string, v ...interface{}) {
	joinFmt := "WebRTCHandler(%p): " + format

	joinArgs := make([]interface{}, len(v)+1)
	joinArgs[0] = wh
	copy(joinArgs[1:], v)

	log.Printf(joinFmt, joinArgs...)
}
