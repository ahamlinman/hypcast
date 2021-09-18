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

		var (
			code         int
			body         interface{}
			params, perr = readRPCParams(r)
		)
		if perr == nil {
			code, body = handler(params)
		} else {
			code, body = errorHTTPCode(perr), perr
		}

		if berr, ok := body.(error); ok {
			body = struct{ Error string }{berr.Error()}
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

type httpError struct {
	HTTPCode int
	Message  string
}

func (h httpError) Error() string { return h.Message }

var (
	errBodyTooLarge    = httpError{http.StatusRequestEntityTooLarge, "RPC body exceeded maximum size"}
	errInvalidBodyType = httpError{http.StatusUnsupportedMediaType, "must have Content-Type: application/json"}
	errInvalidBody     = httpError{http.StatusBadRequest, "failed to parse RPC body as JSON"}
)

func readRPCParams(r *http.Request) (rpcParams, error) {
	const maxBodySize = 1024
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

	var params rpcParams
	err = json.Unmarshal(body.Bytes(), &params)
	if err != nil {
		return nil, errInvalidBody
	}

	return params, nil
}

func errorHTTPCode(err error) int {
	var herr httpError
	if errors.As(err, &herr) {
		return herr.HTTPCode
	}
	return http.StatusInternalServerError
}
