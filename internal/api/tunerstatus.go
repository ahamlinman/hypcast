package api

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
	"github.com/ahamlinman/hypcast/internal/watch"
)

type tunerStatusHandler struct {
	tuner     *tuner.Tuner
	socket    *websocket.Conn
	watch     watch.Watch
	ctx       context.Context
	shutdown  context.CancelCauseFunc
	waitGroup sync.WaitGroup
}

func (h *Handler) handleSocketTunerStatus(w http.ResponseWriter, r *http.Request) {
	tsh := &tunerStatusHandler{tuner: h.tuner}
	tsh.ctx, tsh.shutdown = context.WithCancelCause(context.Background())
	tsh.ServeHTTP(w, r)
}

func (tsh *tunerStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tsh.logf("Starting new connection")
	defer func() {
		tsh.waitForCleanup()
		tsh.logf("Finished with error: %v", tsh.ctx.Err())
	}()

	var err error
	tsh.socket, err = websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer tsh.socket.Close(websocket.StatusGoingAway, "shut down by server")

	tsh.waitGroup.Add(1)
	go func() {
		defer tsh.waitGroup.Done()
		tsh.drainClient()
	}()

	tsh.watch = tsh.tuner.WatchStatus(tsh.sendNewTunerStatus)
	defer tsh.watch.Cancel()

	<-tsh.ctx.Done()
}

func (tsh *tunerStatusHandler) sendNewTunerStatus(s tuner.Status) {
	tsh.logf("Received tuner status: %v", s)
	msg := tsh.mapTunerStatusToMessage(s)
	if err := wsjson.Write(tsh.ctx, tsh.socket, msg); err != nil {
		tsh.shutdown(err)
	}
}

func (tsh *tunerStatusHandler) drainClient() (err error) {
	defer func() { tsh.shutdown(err) }()
	for {
		_, reader, err := tsh.socket.Reader(tsh.ctx)
		switch {
		case websocket.CloseStatus(err) == websocket.StatusGoingAway:
			return nil
		case err != nil:
			return err
		default:
			if _, err := io.Copy(io.Discard, reader); err != nil {
				return err
			}
		}
	}
}

func (tsh *tunerStatusHandler) waitForCleanup() {
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

func (tsh *tunerStatusHandler) logf(format string, v ...any) {
	joinFmt := "TunerStatusHandler(%p): " + format
	joinArgs := make([]any, len(v)+1)
	joinArgs[0] = tsh
	copy(joinArgs[1:], v)
	log.Printf(joinFmt, joinArgs...)
}
