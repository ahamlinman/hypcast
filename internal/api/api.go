package api

import (
	"encoding/json"
	"errors"
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

	h.mux.HandleFunc("/api/sockets/rtc-signaling", h.handleSocketRTCSignaling)
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

func (h *Handler) handleRPCStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := h.tuner.Stop(); err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleRPCTune(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var params struct{ ChannelName string }
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil || params.ChannelName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.tuner.Tune(params.ChannelName)

	if err != nil {
		w.Header().Add("Content-Type", "application/json")

		switch {
		case errors.Is(err, tuner.ErrChannelNotFound):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleSocketRTCSignaling(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleSocketTunerStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
