package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/echowang1/agent-vault/internal/api"
	"github.com/echowang1/agent-vault/internal/config"
	"github.com/echowang1/agent-vault/internal/policy"
	"github.com/echowang1/agent-vault/internal/storage"
	"github.com/echowang1/agent-vault/internal/tss"
	"github.com/gin-gonic/gin"
)

func newServer(cfg *config.Config) (*gin.Engine, error) {
	encryptor, err := storage.NewAES256GCMEncryptorFromEnv()
	if err != nil {
		return nil, err
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/mpc-wallet.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	sqliteStore, err := storage.NewSQLiteStorage(dbPath, encryptor)
	if err != nil {
		return nil, err
	}

	keyGen, err := tss.NewKeyGeneratorWithStorage(sqliteStore)
	if err != nil {
		return nil, err
	}

	signer, err := tss.NewSigner(keyGen.(tss.ShardStorage))
	if err != nil {
		return nil, err
	}

	policyStore := policy.NewSQLiteStorage(sqliteStore.DB())
	policyEngine, err := policy.NewPolicyEngine(policyStore)
	if err != nil {
		return nil, err
	}

	handler := api.NewWalletHandler(keyGen, signer, sqliteStore, policyEngine)
	router := gin.New()
	api.RegisterRoutes(router, handler, cfg.APIKeys)
	return router, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	router, err := newServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Printf("MPC Wallet Server v0.1.0 listening on %s", addr)
	log.Fatal(router.Run(addr))
}
