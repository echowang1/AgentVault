# Task 004: HTTP Server + 基础路由

**负责人**: Codex
**审核人**: Claude
**预计时间**: 3-4 小时
**依赖**: Task 002, 003

---

## 功能要求

### 必须实现 (Must Have)
- [ ] 使用 `gin-gonic/gin` 框架
- [ ] 健康检查端点 `GET /health`
- [ ] 创建钱包 `POST /api/v1/wallet/create`
- [ ] 签名交易 `POST /api/v1/wallet/sign`
- [ ] 查询钱包 `GET /api/v1/wallet/:address`
- [ ] CORS 中间件
- [ ] API Key 认证中间件
- [ ] 请求日志中间件
- [ ] 错误处理统一格式

### 建议实现 (Should Have)
- [ ] 请求 ID 追踪
- [ ] 指标端点 `/metrics`
- [ ] 优雅关闭

---

## API 设计

### 健康检查

```http
GET /health

响应:
{
  "status": "ok",
  "version": "0.1.0",
  "timestamp": "2026-02-21T10:00:00Z"
}
```

### 创建钱包

```http
POST /api/v1/wallet/create
Authorization: Bearer <API_KEY>
Content-Type: application/json

请求体: (可选)
{
  "chain_id": "1"  // 默认 1 (Ethereum)
}

响应:
{
  "success": true,
  "data": {
    "address": "0x...",
    "public_key": "0x...",
    "shard1": "base64...",
    "shard2_id": "uuid-..."
  }
}
```

### 签名交易

```http
POST /api/v1/wallet/sign
Authorization: Bearer <API_KEY>
Content-Type: application/json

请求:
{
  "address": "0x...",
  "message_hash": "0x...",  // 32 字节，十六进制
  "shard1": "base64..."
}

响应:
{
  "success": true,
  "data": {
    "signature": "0x...",  // 130 字符 (0x + r64 + s64 + v2)
    "r": "0x...",
    "s": "0x...",
    "v": 28
  }
}
```

### 查询钱包

```http
GET /api/v1/wallet/:address
Authorization: Bearer <API_KEY>

响应:
{
  "success": true,
  "data": {
    "address": "0x...",
    "public_key": "0x...",
    "created_at": "2026-02-21T10:00:00Z"
  }
}
```

### 错误响应格式

```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "详细错误信息",
    "details": {}
  }
}
```

---

## 接口定义

```go
// server/internal/api/handler.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// WalletHandler 钱包处理器
type WalletHandler struct {
    keyGen   tss.KeyGenerator
    signer   tss.Signer
    storage  storage.ShardStorage
}

// NewWalletHandler 创建处理器
func NewWalletHandler(
    keyGen tss.KeyGenerator,
    signer tss.Signer,
    storage storage.ShardStorage,
) *WalletHandler

// RegisterRoutes 注册路由
func (h *WalletHandler) RegisterRoutes(r *gin.Engine)
```

---

## 中间件要求

### API Key 认证

```go
// server/internal/api/middleware.go
package api

import "github.com/gin-gonic/gin"

// APIKeyAuth API Key 认证中间件
func APIKeyAuth(validKeys map[string]bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("Authorization")
        if apiKey == "" {
            c.JSON(401, gin.H{"error": "missing API key"})
            c.Abort()
            return
        }

        // 提取 Bearer token
        if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
            apiKey = apiKey[7:]
        }

        if !validKeys[apiKey] {
            c.JSON(401, gin.H{"error": "invalid API key"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### CORS

```go
// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

---

## 测试用例

### 单元测试

```go
// server/internal/api/handler_test.go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
    router := setupTestRouter()

    req, _ := http.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "ok")
}

func TestCreateWallet_Success(t *testing.T) {
    router := setupTestRouter()

    body := map[string]interface{}{
        "chain_id": "1",
    }
    jsonBody, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", "/api/v1/wallet/create", bytes.NewReader(jsonBody))
    req.Header.Set("Authorization", "Bearer test-api-key")
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)

    assert.True(t, resp["success"].(bool))
    data := resp["data"].(map[string]interface{})
    assert.Regexp(t, "^0x[a-fA-F0-9]{40}$", data["address"])
    assert.NotEmpty(t, data["shard1"])
    assert.NotEmpty(t, data["shard2_id"])
}

func TestCreateWallet_NoAPIKey(t *testing.T) {
    router := setupTestRouter()

    req, _ := http.NewRequest("POST", "/api/v1/wallet/create", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
}

func TestSign_Success(t *testing.T) {
    // 1. 先创建钱包
    // 2. 然后签名
    // ...
}
```

---

## 配置要求

```go
// server/internal/config/config.go
package config

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    APIKeys  map[string]bool `mapstructure:"api_keys"`
}

type ServerConfig struct {
    Host            string `mapstructure:"host"`
    Port            int    `mapstructure:"port"`
    ReadTimeout     int    `mapstructure:"read_timeout"`
    WriteTimeout    int    `mapstructure:"write_timeout"`
    ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// Load 加载配置
func Load() (*Config, error)
```

---

## 文件结构

```
server/internal/api/
├── handler.go              # 处理器
├── middleware.go           # 中间件
├── routes.go               # 路由注册
├── response.go             # 响应格式
├── handler_test.go         # 单元测试
└── errors.go               # 错误定义
```

---

## 完成标志

### 功能验证
- [ ] 所有 API 端点可访问
- [ ] API Key 认证正常工作
- [ ] 错误响应格式统一
- [ ] 所有测试通过

### 代码质量
- [ ] `go test ./...` 通过
- [ ] `golangci-lint run` 通过
- [ ] 测试覆盖率 > 70%

### 集成验证
- [ ] 可以用 curl 测试所有端点
- [ ] 提供测试脚本

---

## 验证命令

```bash
# 健康检查
curl http://localhost:8080/health

# 创建钱包
curl -X POST http://localhost:8080/api/v1/wallet/create \
  -H "Authorization: Bearer test-key" \
  -H "Content-Type: application/json"

# 签名
curl -X POST http://localhost:8080/api/v1/wallet/sign \
  -H "Authorization: Bearer test-key" \
  -H "Content-Type: application/json" \
  -d '{"address":"0x...","message_hash":"0x...","shard1":"..."}'
```

---

## 下一步

完成后，可以开始 **Task 005: Shard 2 存储层**
