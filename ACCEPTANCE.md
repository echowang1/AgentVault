# Task 002: TSS 密钥生成核心功能

**负责人**: Codex
**审核人**: Claude
**预计时间**: 4-6 小时
**依赖**: Task 001

---

## 功能要求

### 必须实现 (Must Have)
- [ ] 使用 `github.com/bnb-chain/tss-lib` 实现 2-of-2 ECDSA 密钥生成
- [ ] 返回以太坊地址
- [ ] 返回 Shard 1 (给 Agent)
- [ ] 返回 Shard 2 (存储在服务端)
- [ ] 完整的单元测试
- [ ] 集成测试验证签名可用

### 建议实现 (Should Have)
- [ ] 支持自定义曲线 (secp256k1)
- [ ] 密钥生成进度回调
- [ ] 性能优化（预计算）

---

## 接口定义

```go
// server/internal/tss/keygen.go
package tss

import (
    "context"
    "math/big"
)

// KeyGenResult 密钥生成结果
type KeyGenResult struct {
    // Address 以太坊地址（0x开头，42字符）
    Address string `json:"address"`

    // PublicKey 完整公钥（未压缩格式）
    PublicKey string `json:"public_key"`

    // Shard1 密钥碎片 1（base64 编码）
    // Agent 持有，需要安全存储
    Shard1 string `json:"shard1"`

    // Shard2ID 服务端存储的碎片 ID
    // 用于后续签名时检索
    Shard2ID string `json:"shard2_id"`

    // ChainID 链 ID (默认 1 for Ethereum)
    ChainID *big.Int `json:"chain_id"`
}

// KeyGenerateProgress 密钥生成进度
type KeyGenerateProgress struct {
    Step    string `json:"step"`     // 当前步骤
    Percent int    `json:"percent"`  // 进度百分比
}

// ProgressCallback 进度回调函数
type ProgressCallback func(progress KeyGenerateProgress)

// KeyGenerator 密钥生成器接口
type KeyGenerator interface {
    // GenerateKey 生成新的 2-of-2 TSS 密钥对
    GenerateKey(ctx context.Context) (*KeyGenResult, error)

    // GenerateKeyWithProgress 带进度的密钥生成
    GenerateKeyWithProgress(
        ctx context.Context,
        callback ProgressCallback,
    ) (*KeyGenResult, error)
}

// NewKeyGenerator 创建密钥生成器
func NewKeyGenerator() (KeyGenerator, error)
```

---

## 技术要求

### 依赖包

```go
// server/go.mod
require (
    github.com/bnb-chain/tss-lib v2.0.2+incompatible
    github.com/ethereum/go-ethereum v0.13.0
    github.com/stretchr/testify v1.9.0
)
```

### TSS 协议

使用 **GG18 协议** (Gennaro-Goldfeder 2018):

```
2-of-2 配置:
├── parties: 2 (两个参与方)
├── threshold: 1 (需要 2 个签名)
└── curve: secp256k1
```

### 密钥分片存储格式

```go
// KeyShareData 密钥分片数据（内部结构，不对外暴露）
type KeyShareData struct {
    // ShareID 分片唯一 ID
    ShareID string `json:"share_id"`

    // PartyID 参与方 ID
    PartyID string `json:"party_id"`

    // Xi 私钥分片（大整数）
    Xi *big.Int `json:"-"`

    // PublicKey 公钥点
    PublicKey *ecdsa.PublicKey `json:"-"`

    // CreatedAt 创建时间
    CreatedAt time.Time `json:"created_at"`
}
```

---

## 测试用例

### 单元测试

```go
// server/internal/tss/keygen_test.go
package tss

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewKeyGenerator(t *testing.T) {
    keygen, err := NewKeyGenerator()
    require.NoError(t, err)
    assert.NotNil(t, keygen)
}

func TestGenerateKey_ValidResult(t *testing.T) {
    keygen, err := NewKeyGenerator()
    require.NoError(t, err)

    result, err := keygen.GenerateKey(context.Background())
    require.NoError(t, err)

    // 验证地址格式
    assert.Regexp(t, "^0x[a-fA-F0-9]{40}$", result.Address)

    // 验证公钥
    assert.NotEmpty(t, result.PublicKey)

    // 验证 Shard 1
    assert.NotEmpty(t, result.Shard1)
    shard1Bytes, err := base64.StdEncoding.DecodeString(result.Shard1)
    require.NoError(t, err)
    assert.NotEmpty(t, shard1Bytes)

    // 验证 Shard 2 ID
    assert.NotEmpty(t, result.Shard2ID)
}

func TestGenerateKey_Deterministic(t *testing.T) {
    // 相同的随机种子应该生成相同的密钥
    t.Skip("需要先确定随机种子策略")
}

func TestGenerateKey_AddrressChecksum(t *testing.T) {
    keygen, err := NewKeyGenerator()
    require.NoError(t, err)

    result, err := keygen.GenerateKey(context.Background())
    require.NoError(t, err)

    // 验证 EIP-55 checksum
    assert.Equal(t, result.Address, toChecksumAddress(result.Address))
}
```

### 集成测试

```go
// server/internal/tss/keygen_integration_test.go
package tss

import (
    "context"
    "testing"
    "github.com/ethereum/go-ethereum/crypto"
)

func TestKeyGenAndSignIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }

    // 1. 生成密钥
    keygen, err := NewKeyGenerator()
    require.NoError(t, err)

    result, err := keygen.GenerateKey(context.Background())
    require.NoError(t, err)

    // 2. 解码 Shard 1
    shard1Bytes, err := base64.StdEncoding.DecodeString(result.Shard1)
    require.NoError(t, err)

    // 3. 从存储加载 Shard 2 (模拟)
    shard2Bytes, err := loadShard2(result.Shard2ID)
    require.NoError(t, err)

    // 4. 验证两个分片可以恢复私钥并签名
    // (这个测试需要签名模块，先占位)
    t.Skip("等待签名模块实现")
}
```

---

## 性能要求

- [ ] 密钥生成在 10 秒内完成（冷启动）
- [ ] 密钥生成在 3 秒内完成（预计算后）
- [ ] 内存占用不超过 100MB

---

## 安全要求

- [ ] Shard 1 返回前 base64 编码
- [ ] Shard 2 不暴露在 API 响应中
- [ ] 私钥分片永不序列化为 JSON
- [ ] 使用 crypto/rand 生成随机数
- [ ] 错误信息不泄露密钥数据

---

## 错误处理

```go
// 定义错误类型
var (
    ErrKeyGenFailed      = errors.New("密钥生成失败")
    ErrInvalidPartyCount = errors.New("参与方数量必须为 2")
    ErrContextCanceled   = errors.New("操作被取消")
)
```

---

## 文件结构

```
server/internal/tss/
├── keygen.go              # 主实现
├── keygen_test.go         # 单元测试
├── keygen_integration_test.go  # 集成测试
├── types.go               # 共享类型定义
└── precompute.go          # 预计算优化（可选）
```

---

## 实现提示

### 1. 预计算优化

```go
// PreParams 预计算参数（耗时操作）
type PreParams struct {
    // tss-lib 的预计算参数
}

// GeneratePreParams 生成预计算参数
// 建议在服务启动时调用
func GeneratePreParams(timeout time.Duration) (*PreParams, error) {
    // ...
}
```

### 2. 地址计算

```go
// PublicKeyToAddress 将公钥转换为以太坊地址
func PublicKeyToAddress(pubKey *ecdsa.PublicKey) string {
    // 1. Keccak256(公钥无前缀 04)
    // 2. 取后 20 字节
    // 3. 添加 0x 前缀
    // 4. EIP-55 checksum
}
```

### 3. 分片序列化

```go
// MarshalShare 序列化密钥分片
func MarshalShare(share *KeyShareData) (string, error) {
    // 使用 protobuf 或 json
    // 加密后再 base64
}

// UnmarshalShare 反序列化密钥分片
func UnmarshalShare(data string) (*KeyShareData, error)
```

---

## 完成标志

### 功能验证
- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 生成的地址可以被 ethers.js 验证

### 代码质量
- [ ] `go test ./...` 通过
- [ ] `golangci-lint run` 通过
- [ ] 测试覆盖率 > 80%

### 文档
- [ ] 函数有 godoc 注释
- [ ] 复杂逻辑有行内注释
- [ ] 更新 README.md（如有新增配置）

---

## 验证命令

```bash
# 运行单元测试
cd server && go test -v ./internal/tss/...

# 运行集成测试
cd server && go test -v ./internal/tss/... -run Integration

# 检查覆盖率
cd server && go test -coverprofile=coverage.out ./internal/tss/...

# 本地验证地址（使用 ethers.js 或 cast）
cast balance <生成的地址> --rpc-url https://eth.llamarpc.com
```

---

## 下一步

完成后，可以开始 **Task 003: TSS 签名核心功能**

---

## 参考资料

- [bnb-chain/tss-lib README](https://github.com/bnb-chain/tss-lib)
- [GG18 论文](https://eprint.iacr.org/2019/114.pdf)
- [secp256k1 曲线参数](https://en.bitcoin.it/wiki/Secp256k1)
