package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", payload["status"])
	}

	if payload["version"] != "0.1.0" {
		t.Fatalf("expected version=0.1.0, got %q", payload["version"])
	}
}
