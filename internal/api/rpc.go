package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
)

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
