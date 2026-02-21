package api

import (
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/echowang1/agent-vault/internal/policy"
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
	keyGen       tss.KeyGenerator
	signer       tss.Signer
	walletStore  storage.WalletStorage
	policyEngine policy.PolicyEngine

	mu      sync.RWMutex
	wallets map[string]walletRecord
}

func NewWalletHandler(keyGen tss.KeyGenerator, signer tss.Signer, walletStore storage.WalletStorage, policyEngine policy.PolicyEngine) *WalletHandler {
	return &WalletHandler{
		keyGen:       keyGen,
		signer:       signer,
		walletStore:  walletStore,
		policyEngine: policyEngine,
		wallets:      make(map[string]walletRecord),
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
	To          string `json:"to,omitempty"`
	Value       string `json:"value,omitempty"`
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

	walletID, shard2ID, err := h.resolveWallet(c, req.Address, req.Shard2ID)
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "wallet not found", nil)
		return
	}

	amount, err := parseAmount(req.Value)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid value", nil)
		return
	}

	if h.policyEngine != nil {
		err = h.policyEngine.Check(c.Request.Context(), &policy.SignRequest{
			WalletID:  walletID,
			To:        strings.ToLower(req.To),
			Value:     amount,
			Timestamp: time.Now().UTC(),
		})
		if err != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error(), nil)
			return
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

	if h.policyEngine != nil {
		_ = h.policyEngine.IncrementUsage(c.Request.Context(), walletID, amount, time.Now().UTC())
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

type policyRequestDTO struct {
	SingleTxLimit string   `json:"single_tx_limit"`
	DailyLimit    string   `json:"daily_limit"`
	Whitelist     []string `json:"whitelist"`
	DailyTxLimit  int      `json:"daily_tx_limit"`
	StartTime     string   `json:"start_time,omitempty"`
	EndTime       string   `json:"end_time,omitempty"`
}

func (h *WalletHandler) SetPolicy(c *gin.Context) {
	if h.policyEngine == nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, "policy engine unavailable", nil)
		return
	}

	walletID, _, err := h.resolveWallet(c, c.Param("address"), "")
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "wallet not found", nil)
		return
	}

	var req policyRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body", nil)
		return
	}

	st, err := parseOptionalTime(req.StartTime)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid start_time", nil)
		return
	}
	et, err := parseOptionalTime(req.EndTime)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid end_time", nil)
		return
	}

	p := &policy.Policy{
		WalletID:      walletID,
		SingleTxLimit: parseOptionalBig(req.SingleTxLimit),
		DailyLimit:    parseOptionalBig(req.DailyLimit),
		Whitelist:     req.Whitelist,
		DailyTxLimit:  req.DailyTxLimit,
		StartTime:     st,
		EndTime:       et,
	}
	if err := h.policyEngine.SetPolicy(c.Request.Context(), p); err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, "failed to set policy", nil)
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"wallet_id": walletID})
}

func (h *WalletHandler) GetPolicy(c *gin.Context) {
	if h.policyEngine == nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, "policy engine unavailable", nil)
		return
	}
	walletID, _, err := h.resolveWallet(c, c.Param("address"), "")
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "wallet not found", nil)
		return
	}

	p, err := h.policyEngine.GetPolicy(c.Request.Context(), walletID)
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "policy not found", nil)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"wallet_id":       p.WalletID,
		"single_tx_limit": bigToString(p.SingleTxLimit),
		"daily_limit":     bigToString(p.DailyLimit),
		"whitelist":       p.Whitelist,
		"daily_tx_limit":  p.DailyTxLimit,
		"start_time":      formatTime(p.StartTime),
		"end_time":        formatTime(p.EndTime),
	})
}

func (h *WalletHandler) GetUsage(c *gin.Context) {
	if h.policyEngine == nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, "policy engine unavailable", nil)
		return
	}
	walletID, _, err := h.resolveWallet(c, c.Param("address"), "")
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "wallet not found", nil)
		return
	}

	usage, err := h.policyEngine.GetDailyUsage(c.Request.Context(), walletID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, "failed to get usage", nil)
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{
		"wallet_id":    usage.WalletID,
		"date":         usage.Date,
		"total_amount": usage.TotalAmount.String(),
		"tx_count":     usage.TxCount,
	})
}

func (h *WalletHandler) resolveWallet(c *gin.Context, address, shard2IDOverride string) (walletID string, shard2ID string, err error) {
	if h.walletStore != nil {
		info, e := h.walletStore.GetByAddress(c.Request.Context(), address)
		if e == nil {
			sid := info.Shard2ID
			if shard2IDOverride != "" {
				sid = shard2IDOverride
			}
			return info.ID, sid, nil
		}
	}

	h.mu.RLock()
	rec, ok := h.wallets[strings.ToLower(address)]
	h.mu.RUnlock()
	if !ok {
		return "", "", http.ErrNoLocation
	}
	sid := rec.Shard2ID
	if shard2IDOverride != "" {
		sid = shard2IDOverride
	}
	return rec.Shard2ID, sid, nil
}

func parseAmount(raw string) (*big.Int, error) {
	if strings.TrimSpace(raw) == "" {
		return big.NewInt(0), nil
	}
	v := new(big.Int)
	if _, ok := v.SetString(strings.TrimSpace(raw), 10); !ok {
		return nil, policy.ErrInvalidAmount
	}
	if v.Sign() < 0 {
		return nil, policy.ErrInvalidAmount
	}
	return v, nil
}

func parseOptionalBig(raw string) *big.Int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v := new(big.Int)
	if _, ok := v.SetString(raw, 10); !ok {
		return nil
	}
	return v
}

func parseOptionalTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	t = t.UTC()
	return &t, nil
}

func formatTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func bigToString(v *big.Int) string {
	if v == nil {
		return ""
	}
	return v.String()
}
