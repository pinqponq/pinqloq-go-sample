package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPScenarios(t *testing.T) {
	mux := newMux("public")

	for status := range scenarioMessages {
		status := status
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/demo/http/%d", status), nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != status {
				t.Fatalf("expected status %d, got %d", status, rec.Code)
			}

			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}

			if int(body["scenario"].(float64)) != status {
				t.Fatalf("expected scenario %d, got %v", status, body["scenario"])
			}
		})
	}
}

func TestUnsupportedStatus(t *testing.T) {
	mux := newMux("public")

	req := httptest.NewRequest(http.MethodGet, "/demo/http/999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestServesTestLabPage(t *testing.T) {
	mux := newMux("public")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "pinqloq") {
		t.Fatalf("expected page to mention pinqloq")
	}
}

func TestConfigDoesNotLeakSecretKey(t *testing.T) {
	mux := newMux("public")

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}

	if _, ok := body["configured"]; !ok {
		t.Fatalf("expected a 'configured' field in the response")
	}
	if _, ok := body["secretKey"]; ok {
		t.Fatalf("response must never include a secretKey field")
	}
}

func TestManualEventRejectedBeforeSessionConfigured(t *testing.T) {
	client, _, _, _ := currentSession.read()
	if client != nil {
		t.Skip("a session is configured in this environment")
	}

	mux := newMux("public")

	req := httptest.NewRequest(http.MethodPost, "/demo/manual/information", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestSessionRejectsInvalidInput(t *testing.T) {
	mux := newMux("public")

	cases := []struct {
		name string
		body string
	}{
		{"missing fields", `{"secretKey":"","httpCollection":"a","manualCollection":"b"}`},
		{"identical collections", `{"secretKey":"lgl_x","httpCollection":"same","manualCollection":"same"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rec.Code)
			}
		})
	}
}
