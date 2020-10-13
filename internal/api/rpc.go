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

type rpcHandlerFunc func(params rpcParams) (code int, resp interface{})

func (h *Handler) rpcStop(_ rpcParams) (code int, resp interface{}) {
	if err := h.tuner.Stop(); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusNoContent, nil
}

func (h *Handler) rpcTune(params rpcParams) (code int, resp interface{}) {
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

	default:
		return http.StatusNoContent, nil
	}
}

func handleRPC(handler rpcHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Add("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var (
			code int
			resp interface{}

			params, err = readRPCParams(r)
		)

		switch {
		case errors.Is(err, errBodyTooLarge):
			code, resp = http.StatusRequestEntityTooLarge, err
		case errors.Is(err, errInvalidBodyType):
			code, resp = http.StatusUnsupportedMediaType, err
		case errors.Is(err, errInvalidBody):
			code, resp = http.StatusBadRequest, err
		case err != nil:
			code, resp = http.StatusInternalServerError, err

		default:
			code, resp = handler(params)
		}

		if err, ok := resp.(error); ok {
			resp = struct{ Error string }{err.Error()}
		}

		if resp != nil {
			w.Header().Add("Content-Type", "application/json")
		}
		w.WriteHeader(code)
		if resp != nil {
			json.NewEncoder(w).Encode(resp)
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
