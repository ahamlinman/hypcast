package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
)

type rpcParams map[string]interface{}

type rpcHandlerFunc func(rpcParams) (code int, body interface{})

func (h *Handler) rpcStop(_ rpcParams) (code int, body interface{}) {
	if err := h.tuner.Stop(); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusNoContent, nil
}

func (h *Handler) rpcTune(params rpcParams) (code int, body interface{}) {
	channelName, ok := params["ChannelName"].(string)
	if !ok {
		return http.StatusBadRequest, errors.New("channel name required")
	}

	err := h.tuner.Tune(channelName)
	switch {
	case errors.Is(err, tuner.ErrChannelNotFound):
		return http.StatusBadRequest, err
	case err != nil:
		return http.StatusInternalServerError, err
	}

	return http.StatusNoContent, nil
}

func handleRPC(handler rpcHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Add("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		params, err := readRPCParams(r)

		var (
			code int
			body interface{}
		)
		switch {
		case errors.Is(err, errBodyTooLarge):
			code, body = http.StatusRequestEntityTooLarge, err
		case errors.Is(err, errInvalidBodyType):
			code, body = http.StatusUnsupportedMediaType, err
		case errors.Is(err, errInvalidBody):
			code, body = http.StatusBadRequest, err
		case err != nil:
			code, body = http.StatusInternalServerError, err

		default:
			code, body = handler(params)
		}

		if err, ok := body.(error); ok {
			body = struct{ Error string }{err.Error()}
		}

		if body != nil {
			w.Header().Add("Content-Type", "application/json")
		}
		w.WriteHeader(code)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	})
}

const maxBodySize = 1024

var (
	errBodyTooLarge    = errors.New("RPC body exceeded maximum size")
	errInvalidBodyType = errors.New("must have Content-Type: application/json")
	errInvalidBody     = errors.New("failed to parse RPC body as JSON")
)

func readRPCParams(r *http.Request) (rpcParams, error) {
	var body bytes.Buffer
	n, err := body.ReadFrom(io.LimitReader(r.Body, maxBodySize+1))
	switch {
	case n == 0 || err != nil:
		return nil, err
	case n > maxBodySize:
		return nil, errBodyTooLarge
	}

	if r.Header.Get("Content-Type") != "application/json" {
		return nil, errInvalidBodyType
	}

	var params map[string]interface{}
	err = json.Unmarshal(body.Bytes(), &params)
	if err != nil {
		return nil, errInvalidBody
	}

	return params, nil
}
