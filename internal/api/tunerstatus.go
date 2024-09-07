package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
	"github.com/ahamlinman/hypcast/internal/log"
	"github.com/ahamlinman/hypcast/internal/watch"
)

type TunerStatusHandler struct {
	tuner     *tuner.Tuner
	socket    *websocket.Conn
	watch     watch.Watch
	ctx       context.Context
	shutdown  context.CancelCauseFunc
	waitGroup sync.WaitGroup
}

func (h *Handler) handleSocketTunerStatus(w http.ResponseWriter, r *http.Request) {
	ctx, shutdown := context.WithCancelCause(r.Context())
	tsh := &TunerStatusHandler{
		tuner:    h.tuner,
		ctx:      ctx,
		shutdown: shutdown,
	}
	tsh.ServeHTTP(w, r)
}

func (tsh *TunerStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Tprintf(tsh, "Starting new connection")
	defer func() {
		tsh.waitForCleanup()
		log.Tprintf(tsh, "Connection done: %v", context.Cause(tsh.ctx))
	}()

	var err error
	tsh.socket, err = websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer tsh.socket.Close()

	tsh.waitGroup.Add(1)
	go func() {
		defer tsh.waitGroup.Done()
		tsh.drainClient()
	}()

	tsh.watch = tsh.tuner.WatchStatus(tsh.sendNewTunerStatus)
	defer tsh.watch.Cancel()

	<-tsh.ctx.Done()
}

func (tsh *TunerStatusHandler) sendNewTunerStatus(s tuner.Status) {
	log.Tprintf(tsh, "Received tuner status: %v", s)

	msg := tsh.mapTunerStatusToMessage(s)
	if err := tsh.socket.WriteJSON(msg); err != nil {
		tsh.shutdown(err)
	}
}

func (tsh *TunerStatusHandler) drainClient() {
	// Per https://pkg.go.dev/github.com/gorilla/websocket#hdr-Control_Messages,
	// we have to drain incoming messages ourselves even if we don't care about
	// them.
	for {
		if _, _, err := tsh.socket.NextReader(); err != nil {
			tsh.shutdown(err)
			return
		}
	}
}

func (tsh *TunerStatusHandler) waitForCleanup() {
	if tsh.watch != nil {
		tsh.watch.Wait()
	}
	tsh.waitGroup.Wait()
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

func (tsh *TunerStatusHandler) mapTunerStatusToMessage(s tuner.Status) tunerStatusMsg {
	msg := tunerStatusMsg{
		State:       tunerStateStrings[s.State],
		ChannelName: s.ChannelName,
	}
	if s.Error != nil {
		msg.Error = s.Error.Error()
	}
	return msg
}
