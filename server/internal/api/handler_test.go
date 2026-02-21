package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/echowang1/agent-vault/internal/policy"
	"github.com/echowang1/agent-vault/internal/storage"
	"github.com/echowang1/agent-vault/internal/tss"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	encryptor, err := storage.NewAES256GCMEncryptor(make([]byte, 32))
	require.NoError(t, err)

	sqlStore, err := storage.NewSQLiteStorage(":memory:", encryptor)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlStore.Close() })

	keyGen, err := tss.NewKeyGeneratorWithStorage(sqlStore)
	require.NoError(t, err)

	signer, err := tss.NewSigner(keyGen.(tss.ShardStorage))
	require.NoError(t, err)

	policyStore := policy.NewSQLiteStorage(sqlStore.DB())
	policyEngine, err := policy.NewPolicyEngine(policyStore)
	require.NoError(t, err)

	h := NewWalletHandler(keyGen, signer, sqlStore, policyEngine)
	r := gin.New()
	RegisterRoutes(r, h, map[string]bool{"test-api-key": true})
	return r
}

func TestHealthCheck(t *testing.T) {
	r := setupTestRouter(t)
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"status\":\"ok\"")
}

func TestCreateWallet_Success(t *testing.T) {
	r := setupTestRouter(t)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet/create", bytes.NewReader([]byte(`{"chain_id":"1"}`)))
	req.Header.Set("Authorization", "Bearer test-api-key")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Regexp(t, `^0x[0-9a-fA-F]{40}$`, data["address"])
	assert.NotEmpty(t, data["shard1"])
	assert.NotEmpty(t, data["shard2_id"])
}

func TestCreateWallet_NoAPIKey(t *testing.T) {
	r := setupTestRouter(t)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet/create", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPolicyAndSignFlow(t *testing.T) {
	r := setupTestRouter(t)

	createReq, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet/create", bytes.NewReader([]byte(`{}`)))
	createReq.Header.Set("Authorization", "Bearer test-api-key")
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	require.Equal(t, http.StatusOK, createW.Code)

	var createResp struct {
		Success bool `json:"success"`
		Data    struct {
			Address string `json:"address"`
			Shard1  string `json:"shard1"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))

	putPolicyBody := map[string]interface{}{
		"single_tx_limit": "100",
		"daily_limit":     "1000",
		"whitelist":       []string{"0xabc"},
		"daily_tx_limit":  2,
		"start_time":      time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
		"end_time":        time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	}
	policyBytes, _ := json.Marshal(putPolicyBody)
	policyReq, _ := http.NewRequest(http.MethodPut, "/api/v1/wallet/"+createResp.Data.Address+"/policy", bytes.NewReader(policyBytes))
	policyReq.Header.Set("Authorization", "Bearer test-api-key")
	policyReq.Header.Set("Content-Type", "application/json")
	policyW := httptest.NewRecorder()
	r.ServeHTTP(policyW, policyReq)
	require.Equal(t, http.StatusOK, policyW.Code)

	getPolicyReq, _ := http.NewRequest(http.MethodGet, "/api/v1/wallet/"+createResp.Data.Address+"/policy", nil)
	getPolicyReq.Header.Set("Authorization", "Bearer test-api-key")
	getPolicyW := httptest.NewRecorder()
	r.ServeHTTP(getPolicyW, getPolicyReq)
	require.Equal(t, http.StatusOK, getPolicyW.Code)
	assert.Contains(t, getPolicyW.Body.String(), "single_tx_limit")

	signBadBody := map[string]string{
		"address":      createResp.Data.Address,
		"message_hash": "0x95ad83f5c0e9ceccaf53f989ec3b8f226f97d2bd8717fdad4d2aa5b6b0f7d9b5",
		"shard1":       createResp.Data.Shard1,
		"to":           "0xdef",
		"value":        "10",
	}
	signBadBytes, _ := json.Marshal(signBadBody)
	signBadReq, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet/sign", bytes.NewReader(signBadBytes))
	signBadReq.Header.Set("Authorization", "Bearer test-api-key")
	signBadReq.Header.Set("Content-Type", "application/json")
	signBadW := httptest.NewRecorder()
	r.ServeHTTP(signBadW, signBadReq)
	require.Equal(t, http.StatusBadRequest, signBadW.Code)

	signOkBody := map[string]string{
		"address":      createResp.Data.Address,
		"message_hash": "0x95ad83f5c0e9ceccaf53f989ec3b8f226f97d2bd8717fdad4d2aa5b6b0f7d9b5",
		"shard1":       createResp.Data.Shard1,
		"to":           "0xabc",
		"value":        "10",
	}
	signOkBytes, _ := json.Marshal(signOkBody)
	signOkReq, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet/sign", bytes.NewReader(signOkBytes))
	signOkReq.Header.Set("Authorization", "Bearer test-api-key")
	signOkReq.Header.Set("Content-Type", "application/json")
	signOkW := httptest.NewRecorder()
	r.ServeHTTP(signOkW, signOkReq)
	require.Equal(t, http.StatusOK, signOkW.Code)
	assert.Contains(t, signOkW.Body.String(), "signature")

	usageReq, _ := http.NewRequest(http.MethodGet, "/api/v1/wallet/"+createResp.Data.Address+"/usage", nil)
	usageReq.Header.Set("Authorization", "Bearer test-api-key")
	usageW := httptest.NewRecorder()
	r.ServeHTTP(usageW, usageReq)
	require.Equal(t, http.StatusOK, usageW.Code)
	assert.Contains(t, usageW.Body.String(), "tx_count")
}
