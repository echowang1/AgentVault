package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/echowang1/agent-vault/internal/config"
)

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Println("MPC Wallet Server v0.1.0")
	fmt.Printf("Server starting on %s:%d...\n", cfg.ServerHost, cfg.ServerPort)

	http.HandleFunc("/health", healthHandler)
	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Fatal(http.ListenAndServe(addr, nil))
}
