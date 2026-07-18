package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubDBPing struct {
	err error
}

func (s stubDBPing) Ping(_ context.Context) error {
	return s.err
}

func TestReadyHandlerWithoutDatabase(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	readyHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestReadyHandlerWithDatabaseError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	readyHandler(stubDBPing{err: errors.New("ping failed")}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestReadyHandlerWithHealthyDatabase(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	readyHandler(stubDBPing{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", rec.Body.String())
	}
}

func TestMetricsEndpointReturnsPrometheusMetrics(t *testing.T) {
	handler := NewRouter(Deps{JWTSecret: "test-secret"})

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	handler.ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("expected metrics status %d, got %d", http.StatusOK, metricsRec.Code)
	}
	body := metricsRec.Body.String()
	if !strings.Contains(body, "vault_api_http_requests_total") ||
		!strings.Contains(body, "vault_api_http_request_duration_seconds") {
		t.Fatalf("expected prometheus metrics in body, got %q", body)
	}
}
