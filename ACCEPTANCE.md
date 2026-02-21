# Task 003: TSS 签名核心功能

**负责人**: Codex
**审核人**: Claude
**预计时间**: 4-6 小时
**依赖**: Task 002

---

## 功能要求

### 必须实现 (Must Have)
- [ ] 使用 `github.com/bnb-chain/tss-lib` 实现 2-of-2 ECDSA 签名
- [ ] 接收 Shard 1（从 Agent）和 Shard 2（从存储）
- [ ] 返回标准 ECDSA 签名 (r, s, v)
- [ ] 支持消息签名和交易哈希签名
- [ ] 完整的单元测试
- [ ] 签名可以被 ethers.js 验证

### 建议实现 (Should Have)
- [ ] 批量签名支持（多个消息）
- [ ] 签名进度回调
- [ ] 签名缓存（避免重复签名相同消息）

---

## 接口定义

```go
// server/internal/tss/signing.go
package tss

import (
    "context"
    "math/big"
)

// SignRequest 签名请求
type SignRequest struct {
    // Address 钱包地址
    Address string `json:"address"`

    // MessageHash 要签名的消息哈希（32 字节）
    MessageHash string `json:"message_hash"`

    // Shard1 密钥碎片 1（base64 编码，由 Agent 提供）
    Shard1 string `json:"shard1"`

    // Shard2ID 密钥碎片 2 的存储 ID
    Shard2ID string `json:"shard2_id"`
}

// Signature ECDSA 签名
type Signature struct {
    // R 签名的 r 值（十六进制字符串）
    R string `json:"r"`

    // S 签名的 s 值（十六进制字符串）
    S string `json:"s"`

    // V recovery id (0 或 1)
    V uint8 `json:"v"`

    // 完整签名 (0x 前缀，130 字符)
    // 格式: r(64) + s(64) + v(2)
    FullSignature string `json:"signature"`
}

// SignProgress 签名进度
type SignProgress struct {
    Step    string `json:"step"`
    Percent int    `json:"percent"`
}

// ProgressCallback 进度回调函数
type ProgressCallback func(progress SignProgress)

// Signer 签名器接口
type Signer interface {
    // Sign 签名消息
    Sign(ctx context.Context, req *SignRequest) (*Signature, error)

    // SignWithProgress 带进度的签名
    SignWithProgress(
        ctx context.Context,
        req *SignRequest,
        callback ProgressCallback,
    ) (*Signature, error)

    // SignBatch 批量签名
    SignBatch(
        ctx context.Context,
        reqs []*SignRequest,
    ) ([]*Signature, error)
}

// NewSigner 创建签名器
func NewSigner(shardStorage ShardStorage) (Signer, error)
```

---

## 技术要求

### 签名流程

```
1. 验证输入
   ├── MessageHash 必须是 32 字节
   ├── Address 格式正确
   └── Shard1 解码成功

2. 加载 Shard 2
   ├── 从存储加载
   └── 验证与 Address 匹配

3. 协同签名
   ├── 使用 tss-lib 的 signing 协议
   ├── Shard1 + Shard2 协同计算
   └── 生成 ECDSA 签名

4. 返回结果
   ├── R, S, V
   └── 完整签名字符串
```

### V 值计算

```go
// CalculateV 计算 recovery id
func CalculateV(r, s *big.Int, publicKey *ecdsa.PublicKey, hash []byte) uint8 {
    // 尝试 v = 0 和 v = 1
    // 看哪个能恢复出正确的公钥
    for v := uint8(0); v <= 1; v++ {
        if recoveredPubKey := ec.RecoverToPublic(hash, r, s, v); recoveredPubKey != nil {
            if recoveredPubKey.Equal(publicKey) {
                return v + 27 // 以太坊使用 27/28
            }
        }
    }
    return 0
}
```

---

## 测试用例

### 单元测试

```go
// server/internal/tss/signing_test.go
package tss

import (
    "context"
    "encoding/hex"
    "math/big"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewSigner(t *testing.T) {
    storage := NewMockShardStorage()
    signer, err := NewSigner(storage)
    require.NoError(t, err)
    assert.NotNil(t, signer)
}

func TestSign_ValidSignature(t *testing.T) {
    // 准备：先创建密钥
    keygen, err := NewKeyGenerator()
    require.NoError(t, err)

    keyResult, err := keygen.GenerateKey(context.Background())
    require.NoError(t, err)

    // 存储 Shard 2
    storage := NewMockShardStorage()
    storage.Store(keyResult.Shard2ID, keyResult.Shard2)

    // 创建签名器
    signer, err := NewSigner(storage)
    require.NoError(t, err)

    // 准备签名请求
    messageHash := keccak256Hash([]byte("test message"))
    req := &SignRequest{
        Address:     keyResult.Address,
        MessageHash: hex.EncodeToString(messageHash),
        Shard1:      keyResult.Shard1,
        Shard2ID:    keyResult.Shard2ID,
    }

    // 执行签名
    sig, err := signer.Sign(context.Background(), req)
    require.NoError(t, err)

    // 验证签名格式
    assert.NotEmpty(t, sig.R)
    assert.NotEmpty(t, sig.S)
    assert.Contains(t, []uint8{27, 28}, sig.V)
    assert.Len(t, sig.FullSignature, 132) // 0x + 130 字符

    // 验证签名有效性（使用 go-ethereum）
    rBytes, _ := hex.DecodeString(sig.R)
    sBytes, _ := hex.DecodeString(sig.S)
    rInt := new(big.Int).SetBytes(rBytes)
    sInt := new(big.Int).SetBytes(sBytes)

    pubKey, err := recoverPubKey(messageHash, rInt, sInt, sig.V-27)
    require.NoError(t, err)

    // 比较地址
    recoveredAddr := crypto.PubkeyToAddress(*pubKey)
    assert.Equal(t, keyResult.Address, recoveredAddr.Hex())
}

func TestSign_InvalidHash(t *testing.T) {
    storage := NewMockShardStorage()
    signer, err := NewSigner(storage)
    require.NoError(t, err)

    req := &SignRequest{
        MessageHash: "invalid",
        Shard1:      "base64data",
    }

    _, err = signer.Sign(context.Background(), req)
    assert.Error(t, err)
    assert.Equal(t, ErrInvalidHash, err)
}

func TestSign_ShardNotFound(t *testing.T) {
    storage := NewMockShardStorage()
    signer, err := NewSigner(storage)
    require.NoError(t, err)

    req := &SignRequest{
        Address:     "0x1234567890123456789012345678901234567890",
        MessageHash: hex.EncodeToString(make([]byte, 32)),
        Shard1:      "base64data",
        Shard2ID:    "non-existent",
    }

    _, err = signer.Sign(context.Background(), req)
    assert.Error(t, err)
    assert.Equal(t, ErrShardNotFound, err)
}
```

### 集成测试

```go
// server/internal/tss/signing_integration_test.go
package tss

func TestKeyGenAndSignEndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }

    ctx := context.Background()

    // 1. 生成密钥
    keygen, err := NewKeyGenerator()
    require.NoError(t, err)

    keyResult, err := keygen.GenerateKey(ctx)
    require.NoError(t, err)

    // 2. 存储 Shard 2
    storage := NewMemoryShardStorage()
    err = storage.Store(keyResult.Shard2ID, keyResult.Shard2)
    require.NoError(t, err)

    // 3. 创建签名器并签名
    signer, err := NewSigner(storage)
    require.NoError(t, err)

    message := []byte("Hello, Agent MPC Wallet!")
    messageHash := crypto.Keccak256Hash(message)

    req := &SignRequest{
        Address:     keyResult.Address,
        MessageHash: hex.EncodeToString(messageHash.Bytes()),
        Shard1:      keyResult.Shard1,
        Shard2ID:    keyResult.Shard2ID,
    }

    sig, err := signer.Sign(ctx, req)
    require.NoError(t, err)

    // 4. 验证签名
    rBytes, _ := hex.DecodeString(sig.R)
    sBytes, _ := hex.DecodeString(sig.S)
    rInt := new(big.Int).SetBytes(rBytes)
    sInt := new(big.Int).SetBytes(sBytes)

    pubKeyBytes, err := crypto.Ecrecover(messageHash.Bytes(), rInt, sInt, sig.V-27)
    require.NoError(t, err)

    pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
    require.NoError(t, err)

    recoveredAddr := crypto.PubkeyToAddress(*pubKey)
    assert.Equal(t, keyResult.Address, recoveredAddr.Hex())
}
```

---

## 性能要求

- [ ] 单次签名在 3 秒内完成
- [ ] 批量签名（10 个）在 10 秒内完成
- [ ] 内存占用不超过 50MB

---

## 安全要求

- [ ] 验证 MessageHash 长度（必须是 32 字节）
- [ ] 验证 Shard 1 与 Address 匹配
- [ ] 私钥分片永不暴露在日志中
- [ ] 签名后立即清理内存中的敏感数据
- [ ] 使用 constant-time 比较避免时序攻击

---

## 错误处理

```go
var (
    ErrInvalidHash      = errors.New("消息哈希必须是 32 字节")
    ErrInvalidSignature = errors.New("无效签名")
    ErrShardNotFound    = errors.New("密钥分片未找到")
    ErrShardMismatch    = errors.New("密钥分片不匹配")
    ErrSignFailed       = errors.New("签名失败")
)
```

---

## 文件结构

```
server/internal/tss/
├── signing.go              # 主实现
├── signing_test.go         # 单元测试
├── signing_integration_test.go  # 集成测试
└── mock_storage.go         # Mock 存储（用于测试）
```

---

## 完成标志

### 功能验证
- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 签名可以被 ethers.js 验证

### 代码质量
- [ ] `go test ./...` 通过
- [ ] `golangci-lint run` 通过
- [ ] 测试覆盖率 > 80%

### 验证脚本
- [ ] 提供验证脚本（使用 cast 或 ethers.js）

---

## 验证命令

```bash
# 运行测试
cd server && go test -v ./internal/tss/...

# 使用 ethers.js 验证（在本地测试网）
node scripts/verify-signature.js <address> <message> <signature>

# 或使用 cast
cast verify-signature \
  <address> \
  <message-hash> \
  <signature> \
  --rpc-url https://eth-sepolia.publicnode.com
```

---

## 下一步

完成后，可以开始 **Task 004: HTTP API 服务**

---

## 参考资料

- [bnb-chain/tss-lib signing](https://github.com/bnb-chain/tss-lib/tree/master/ecdsa/signing)
- [EIP-191 签名标准](https://eips.ethereum.org/EIPS/eip-191)
- [EIP-155 交易签名](https://eips.ethereum.org/EIPS/eip-155)
