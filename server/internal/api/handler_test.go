package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/echowang1/agent-vault/internal/tss"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	keyGen, err := tss.NewKeyGenerator()
	require.NoError(t, err)

	signer, err := tss.NewSigner(keyGen.(tss.ShardStorage))
	require.NoError(t, err)

	h := NewWalletHandler(keyGen, signer)
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

func TestSignAndGetWallet_Success(t *testing.T) {
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
			Address  string `json:"address"`
			Shard1   string `json:"shard1"`
			Shard2ID string `json:"shard2_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))

	signBody := map[string]string{
		"address":      createResp.Data.Address,
		"message_hash": "0x95ad83f5c0e9ceccaf53f989ec3b8f226f97d2bd8717fdad4d2aa5b6b0f7d9b5",
		"shard1":       createResp.Data.Shard1,
	}
	bodyBytes, _ := json.Marshal(signBody)
	signReq, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet/sign", bytes.NewReader(bodyBytes))
	signReq.Header.Set("Authorization", "Bearer test-api-key")
	signReq.Header.Set("Content-Type", "application/json")
	signW := httptest.NewRecorder()
	r.ServeHTTP(signW, signReq)
	require.Equal(t, http.StatusOK, signW.Code)
	assert.Contains(t, signW.Body.String(), "signature")

	getReq, _ := http.NewRequest(http.MethodGet, "/api/v1/wallet/"+createResp.Data.Address, nil)
	getReq.Header.Set("Authorization", "Bearer test-api-key")
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	require.Equal(t, http.StatusOK, getW.Code)
	assert.Contains(t, getW.Body.String(), createResp.Data.Address)
}
