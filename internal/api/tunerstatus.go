package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
)

func (h *Handler) handleSocketTunerStatus(w http.ResponseWriter, r *http.Request) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	tsh := &tunerStatusHandler{
		conn:  conn,
		tuner: h.tuner,
	}
	tsh.run()
}

type tunerStatusHandler struct {
	conn       *websocket.Conn
	tuner      *tuner.Tuner
	err        chan error
	readerDone chan struct{}
}

func (tsh *tunerStatusHandler) run() (err error) {
	tsh.logf("Starting new connection")
	defer func() { tsh.logf("Finished with error: %v", err) }()

	tsh.err = make(chan error, 1)
	tsh.readerDone = make(chan struct{})

	go tsh.drainReader()
	defer func() { <-tsh.readerDone }()

	w := tsh.tuner.WatchStatus(tsh.sendNewTunerStatus)
	defer func() {
		w.Cancel()
		w.Wait()
	}()

	return <-tsh.err
}

func (tsh *tunerStatusHandler) sendNewTunerStatus(s tuner.Status) {
	tsh.logf("Received tuner status: %v", s)

	msg := tsh.mapTunerStatusToMessage(s)
	if err := tsh.conn.WriteJSON(msg); err != nil {
		tsh.shutdown(err)
	}
}

func (tsh *tunerStatusHandler) drainReader() {
	defer close(tsh.readerDone)
	// see https://pkg.go.dev/github.com/gorilla/websocket#hdr-Control_Messages
	for {
		if _, _, err := tsh.conn.NextReader(); err != nil {
			tsh.shutdown(err)
			return
		}
	}
}

func (tsh *tunerStatusHandler) shutdown(err error) {
	select {
	case tsh.err <- err:
	default:
	}
}

type tunerStatusMsg struct {
	State       string
	ChannelName string `json:",omitempty"`
	Error       string `json:",omitempty"`
}

var tunerStateStrings = map[tuner.State]string{
	tuner.StateStopped:  "Stopped",
	tuner.StateStarting: "Starting",
	tuner.StatePlaying:  "Playing",
}

func (tsh *tunerStatusHandler) mapTunerStatusToMessage(s tuner.Status) tunerStatusMsg {
	msg := tunerStatusMsg{
		State:       tunerStateStrings[s.State],
		ChannelName: s.ChannelName,
	}
	if s.Error != nil {
		msg.Error = s.Error.Error()
	}
	return msg
}

func (tsh *tunerStatusHandler) logf(format string, v ...interface{}) {
	joinFmt := "TunerStatusHandler(%p): " + format

	joinArgs := make([]interface{}, len(v)+1)
	joinArgs[0] = tsh
	copy(joinArgs[1:], v)

	log.Printf(joinFmt, joinArgs...)
}
