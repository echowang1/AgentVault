package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/echowang1/agent-vault/internal/storage"
	"github.com/echowang1/agent-vault/internal/tss"
	"github.com/gin-gonic/gin"
)

type walletRecord struct {
	Address   string    `json:"address"`
	PublicKey string    `json:"public_key"`
	Shard2ID  string    `json:"shard2_id"`
	CreatedAt time.Time `json:"created_at"`
}

type WalletHandler struct {
	keyGen tss.KeyGenerator
	signer tss.Signer

	walletStore storage.WalletStorage

	mu      sync.RWMutex
	wallets map[string]walletRecord
}

func NewWalletHandler(keyGen tss.KeyGenerator, signer tss.Signer, walletStore storage.WalletStorage) *WalletHandler {
	return &WalletHandler{
		keyGen:      keyGen,
		signer:      signer,
		walletStore: walletStore,
		wallets:     make(map[string]walletRecord),
	}
}

func (h *WalletHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"version":   "0.1.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *WalletHandler) CreateWallet(c *gin.Context) {
	result, err := h.keyGen.GenerateKey(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, "failed to create wallet", nil)
		return
	}

	rec := walletRecord{
		Address:   result.Address,
		PublicKey: result.PublicKey,
		Shard2ID:  result.Shard2ID,
		CreatedAt: time.Now().UTC(),
	}
	if h.walletStore != nil {
		_ = h.walletStore.Create(c.Request.Context(), &storage.WalletInfo{
			ID:        rec.Shard2ID,
			Address:   rec.Address,
			PublicKey: rec.PublicKey,
			Shard2ID:  rec.Shard2ID,
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.CreatedAt,
		})
	}

	h.mu.Lock()
	h.wallets[strings.ToLower(result.Address)] = rec
	h.mu.Unlock()

	respondSuccess(c, http.StatusOK, gin.H{
		"address":    result.Address,
		"public_key": result.PublicKey,
		"shard1":     result.Shard1,
		"shard2_id":  result.Shard2ID,
	})
}

type signRequestDTO struct {
	Address     string `json:"address"`
	MessageHash string `json:"message_hash"`
	Shard1      string `json:"shard1"`
	Shard2ID    string `json:"shard2_id,omitempty"`
}

func (h *WalletHandler) Sign(c *gin.Context) {
	var req signRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body", nil)
		return
	}

	if req.Address == "" || req.MessageHash == "" || req.Shard1 == "" {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "address, message_hash and shard1 are required", nil)
		return
	}

	shard2ID := req.Shard2ID
	if shard2ID == "" {
		if h.walletStore != nil {
			info, err := h.walletStore.GetByAddress(c.Request.Context(), req.Address)
			if err == nil {
				shard2ID = info.Shard2ID
			}
		}
		if shard2ID == "" {
			h.mu.RLock()
			rec, ok := h.wallets[strings.ToLower(req.Address)]
			h.mu.RUnlock()
			if !ok {
				respondError(c, http.StatusNotFound, ErrCodeNotFound, "wallet not found", nil)
				return
			}
			shard2ID = rec.Shard2ID
		}
	}

	sig, err := h.signer.Sign(c.Request.Context(), &tss.SignRequest{
		Address:     req.Address,
		MessageHash: req.MessageHash,
		Shard1:      req.Shard1,
		Shard2ID:    shard2ID,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error(), nil)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"signature": sig.FullSignature,
		"r":         "0x" + sig.R,
		"s":         "0x" + sig.S,
		"v":         sig.V,
	})
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if h.walletStore != nil {
		info, err := h.walletStore.GetByAddress(c.Request.Context(), c.Param("address"))
		if err == nil {
			respondSuccess(c, http.StatusOK, gin.H{
				"address":    info.Address,
				"public_key": info.PublicKey,
				"created_at": info.CreatedAt.Format(time.RFC3339),
			})
			return
		}
	}

	h.mu.RLock()
	rec, ok := h.wallets[address]
	h.mu.RUnlock()
	if !ok {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "wallet not found", nil)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"address":    rec.Address,
		"public_key": rec.PublicKey,
		"created_at": rec.CreatedAt.Format(time.RFC3339),
	})
}
