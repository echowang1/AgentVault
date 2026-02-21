package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/echowang1/agent-vault/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	r, err := newServer(&config.Config{APIKeys: map[string]bool{"test-api-key": true}})
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}
