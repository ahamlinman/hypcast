package rpc_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ahamlinman/hypcast/internal/api/rpc"
)

func Example() {
	setEnabled := func(enabled bool) { /* do something useful */ }

	mux := http.NewServeMux()
	mux.Handle("/rpc/setstatus",
		rpc.HTTPHandler(func(params struct{ Enabled *bool }) (code int, body any) {
			if params.Enabled == nil {
				return http.StatusBadRequest, errors.New(`missing "Enabled" parameter`)
			}

			setEnabled(*params.Enabled)
			return http.StatusNoContent, nil
		}))

	req := httptest.NewRequest(
		http.MethodPost, "/rpc/setstatus", strings.NewReader(`{"Enabled": true}`),
	)
	req.Header.Add("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	fmt.Println(resp.Result().StatusCode)
	// Output: 204
}

func TestRPC(t *testing.T) {
	originalMaxSize := rpc.MaxRequestBodySize
	rpc.MaxRequestBodySize = 32
	defer func() { rpc.MaxRequestBodySize = originalMaxSize }()

	handler := func(_ struct{}) (code int, body any) {
		return http.StatusNoContent, nil
	}

	jsonHeaders := http.Header{"Content-Type": {"application/json"}}

	testCases := []struct {
		Description string
		Method      string
		Body        string
		Headers     http.Header
		WantCode    int
		WantHeaders http.Header
	}{
		{
			Description: "empty body",
			WantCode:    http.StatusNoContent,
		},
		{
			Description: "body with maximum length",
			Body:        `{"Message":"123456789012345678"}`,
			Headers:     jsonHeaders,
			WantCode:    http.StatusNoContent,
		},
		{
			Description: "body too long by 1 character",
			Body:        `{"Message":"1234567890123456789"}`,
			Headers:     jsonHeaders,
			WantCode:    http.StatusRequestEntityTooLarge,
			WantHeaders: jsonHeaders,
		},
		{
			Description: "missing Content-Type header",
			Body:        `{"Valid":false}`,
			WantCode:    http.StatusUnsupportedMediaType,
			WantHeaders: jsonHeaders,
		},
		{
			Description: "invalid JSON body",
			Body:        `{{{]]]`,
			Headers:     jsonHeaders,
			WantCode:    http.StatusBadRequest,
			WantHeaders: jsonHeaders,
		},
		{
			Description: "invalid HTTP method",
			Method:      http.MethodGet,
			WantCode:    http.StatusMethodNotAllowed,
			WantHeaders: http.Header{"Allow": {http.MethodPost}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			method := tc.Method
			if method == "" {
				method = http.MethodPost
			}

			req := httptest.NewRequest(method, "/", strings.NewReader(tc.Body))
			req.Header = tc.Headers

			resp := httptest.NewRecorder()
			rpc.HTTPHandler(handler).ServeHTTP(resp, req)

			if resp.Result().StatusCode != tc.WantCode {
				t.Errorf("wrong status: got %d, want %d", resp.Result().StatusCode, tc.WantCode)
			}

			diff := cmp.Diff(tc.WantHeaders, resp.Result().Header, cmpopts.EquateEmpty())
			if diff != "" {
				t.Errorf("wrong headers (-want +got)\n%s", diff)
			}

			body := string(must(io.ReadAll(resp.Result().Body)))
			if body != "" {
				t.Logf("response body: %s", body)
			}
		})
	}
}

func must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}
