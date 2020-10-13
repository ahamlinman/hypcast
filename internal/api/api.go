package api

import (
	"encoding/json"
	"net/http"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
)

// Handler serves the Hypcast API for a single tuner.
type Handler struct {
	mux   *http.ServeMux
	tuner *tuner.Tuner
}

// NewHandler creates a Handler serving the Hypcast API for tuner.
func NewHandler(tuner *tuner.Tuner) *Handler {
	h := &Handler{
		mux:   http.NewServeMux(),
		tuner: tuner,
	}

	h.mux.HandleFunc("/api/config/channels", h.handleConfigChannels)

	h.mux.HandleFunc("/api/rpc/stop", h.handleRPCStop)
	h.mux.HandleFunc("/api/rpc/tune", h.handleRPCTune)

	h.mux.HandleFunc("/api/sockets/webrtc-signaling", h.handleSocketWebRTCSignaling)
	h.mux.HandleFunc("/api/sockets/tuner-status", h.handleSocketTunerStatus)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleConfigChannels(w http.ResponseWriter, r *http.Request) {
	var (
		channels = h.tuner.Channels()
		names    = make([]string, len(channels))
	)
	for i, ch := range channels {
		names[i] = ch.Name
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func (h *Handler) handleSocketWebRTCSignaling(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleSocketTunerStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
