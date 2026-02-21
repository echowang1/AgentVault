package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/echowang1/agent-vault/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	key := make([]byte, 32)
	t.Setenv("SHARD_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(key))
	t.Setenv("DB_PATH", filepath.Join(t.TempDir(), "test.db"))
	_ = os.Unsetenv("DB_TYPE")

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
