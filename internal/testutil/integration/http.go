//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type JSONRequest struct {
	Method string
	Path   string
	Body   any
	Token  string
}

type JSONResponse struct {
	Status int
	Body   []byte
}

func DoJSON(t *testing.T, handler http.Handler, req JSONRequest) JSONResponse {
	t.Helper()

	var body io.Reader
	if req.Body != nil {
		payload, err := json.Marshal(req.Body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		body = bytes.NewReader(payload)
	}

	httpReq := httptest.NewRequest(req.Method, req.Path, body)
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if req.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httpReq)

	respBody, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	_ = rec.Result().Body.Close()

	return JSONResponse{
		Status: rec.Code,
		Body:   respBody,
	}
}

func DecodeJSON[T any](t *testing.T, resp JSONResponse, dst *T) {
	t.Helper()
	if err := json.Unmarshal(resp.Body, dst); err != nil {
		t.Fatalf("decode json response %q: %v", string(resp.Body), err)
	}
}

func ValidEncryptedBlob(extra ...byte) []byte {
	data := []byte{0x01, 0xAA, 0xBB}
	return append(data, extra...)
}
