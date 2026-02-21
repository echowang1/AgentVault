# Task 006: 策略引擎 (基础版)

**负责人**: Codex
**审核人**: Claude
**预计时间**: 3-4 小时
**依赖**: Task 004

---

## 功能要求

### 必须实现 (Must Have)
- [ ] 单笔金额上限检查
- [ ] 每日金额上限检查
- [ ] 地址白名单检查
- [ ] 策略配置管理
- [ ] 策略持久化存储
- [ ] 完整的单元测试

### 建议实现 (Should Have)
- [ ] 每日交易笔数上限
- [ ] 时间窗口限制
- [ ] 策略优先级
- [ ] 策略版本管理

---

## 接口定义

```go
// server/internal/policy/engine.go
package policy

import (
    "context"
    "math/big"
    "time"
)

// Policy 策略定义
type Policy struct {
    // ID 策略唯一 ID
    ID string `json:"id"`

    // WalletID 应用此策略的钱包 ID
    WalletID string `json:"wallet_id"`

    // SingleTxLimit 单笔交易限额（wei）
    // nil 表示不限制
    SingleTxLimit *big.Int `json:"single_tx_limit"`

    // DailyLimit 每日累计限额（wei）
    // nil 表示不限制
    DailyLimit *big.Int `json:"daily_limit"`

    // Whitelist 白名单地址（允许的目标地址）
    // 空列表表示不限制
    Whitelist []string `json:"whitelist"`

    // DailyTxLimit 每日交易笔数上限
    // 0 表示不限制
    DailyTxLimit int `json:"daily_tx_limit"`

    // StartTime 开始时间（可选）
    // 只在此时间后允许交易
    StartTime *time.Time `json:"start_time"`

    // EndTime 结束时间（可选）
    // 只在此时间前允许交易
    EndTime *time.Time `json:"end_time"`

    // CreatedAt 创建时间
    CreatedAt time.Time `json:"created_at"`

    // UpdatedAt 更新时间
    UpdatedAt time.Time `json:"updated_at"`
}

// SignRequest 签名请求（策略检查用）
type SignRequest struct {
    // WalletID 钱包 ID
    WalletID string

    // To 目标地址
    To string

    // Value 交易金额（wei）
    Value *big.Int

    // Timestamp 请求时间戳
    Timestamp time.Time
}

// PolicyEngine 策略引擎接口
type PolicyEngine interface {
    // Check 检查请求是否符合策略
    // 返回错误表示不符合策略
    Check(ctx context.Context, req *SignRequest) error

    // SetPolicy 设置钱包策略
    SetPolicy(ctx context.Context, policy *Policy) error

    // GetPolicy 获取钱包策略
    GetPolicy(ctx context.Context, walletID string) (*Policy, error)

    // DeletePolicy 删除钱包策略
    DeletePolicy(ctx context.Context, walletID string) error

    // GetDailyUsage 获取今日已使用额度
    GetDailyUsage(ctx context.Context, walletID string) (*DailyUsage, error)
}

// DailyUsage 每日使用情况
type DailyUsage struct {
    // WalletID 钱包 ID
    WalletID string `json:"wallet_id"`

    // Date 日期
    Date string `json:"date"` // YYYY-MM-DD

    // TotalAmount 已交易总额（wei）
    TotalAmount *big.Int `json:"total_amount"`

    // TxCount 交易笔数
    TxCount int `json:"tx_count"`
}

// NewPolicyEngine 创建策略引擎
func NewPolicyEngine(storage Storage) (PolicyEngine, error)
```

---

## 策略检查逻辑

```go
// server/internal/policy/engine.go

func (e *Engine) Check(ctx context.Context, req *SignRequest) error {
    // 1. 获取策略
    policy, err := e.storage.GetPolicy(ctx, req.WalletID)
    if err != nil {
        return ErrPolicyNotFound
    }

    // 2. 单笔限额检查
    if policy.SingleTxLimit != nil {
        if req.Value.Cmp(policy.SingleTxLimit) > 0 {
            return ErrExceedsSingleTxLimit
        }
    }

    // 3. 白名单检查
    if len(policy.Whitelist) > 0 {
        if !e.isWhitelisted(policy.Whitelist, req.To) {
            return ErrAddressNotWhitelisted
        }
    }

    // 4. 时间窗口检查
    if policy.StartTime != nil && req.Timestamp.Before(*policy.StartTime) {
        return ErrOutsideTimeWindow
    }
    if policy.EndTime != nil && req.Timestamp.After(*policy.EndTime) {
        return ErrOutsideTimeWindow
    }

    // 5. 每日限额检查
    if policy.DailyLimit != nil || policy.DailyTxLimit > 0 {
        usage, err := e.storage.GetDailyUsage(ctx, req.WalletID, time.Now())
        if err != nil {
            return err
        }

        // 检查金额
        if policy.DailyLimit != nil {
            newTotal := new(big.Int).Add(usage.TotalAmount, req.Value)
            if newTotal.Cmp(policy.DailyLimit) > 0 {
                return ErrExceedsDailyLimit
            }
        }

        // 检查笔数
        if policy.DailyTxLimit > 0 && usage.TxCount >= policy.DailyTxLimit {
            return ErrExceedsDailyTxLimit
        }
    }

    return nil
}

// isWhitelisted 检查地址是否在白名单中
func (e *Engine) isWhitelisted(whitelist []string, address string) bool {
    normalized := strings.ToLower(address)
    for _, addr := range whitelist {
        if strings.ToLower(addr) == normalized {
            return true
        }
    }
    return false
}
```

---

## 错误定义

```go
var (
    ErrPolicyNotFound        = errors.New("策略未找到")
    ErrExceedsSingleTxLimit  = errors.New("超过单笔交易限额")
    ErrExceedsDailyLimit     = errors.New("超过每日交易限额")
    ErrExceedsDailyTxLimit   = errors.New("超过每日交易笔数限额")
    ErrAddressNotWhitelisted = errors.New("目标地址不在白名单中")
    ErrOutsideTimeWindow     = errors.New("不在允许的时间窗口内")
)
```

---

## 存储接口

```go
// server/internal/policy/storage.go
package policy

import (
    "context"
    "time"
)

// Storage 策略存储接口
type Storage interface {
    // SetPolicy 保存策略
    SetPolicy(ctx context.Context, policy *Policy) error

    // GetPolicy 获取策略
    GetPolicy(ctx context.Context, walletID string) (*Policy, error)

    // DeletePolicy 删除策略
    DeletePolicy(ctx context.Context, walletID string) error

    // GetDailyUsage 获取每日使用情况
    GetDailyUsage(ctx context.Context, walletID string, date time.Time) (*DailyUsage, error)

    // IncrementUsage 增加使用量
    IncrementUsage(ctx context.Context, walletID string, date time.Time, amount *big.Int) error

    // ResetDailyUsage 重置每日使用（用于新的一天）
    ResetDailyUsage(ctx context.Context, walletID string, date time.Time) error
}
```

---

## 测试用例

```go
// server/internal/policy/engine_test.go
package policy

import (
    "context"
    "math/big"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCheck_SingleTxLimit(t *testing.T) {
    engine := setupTestEngine(t)
    ctx := context.Background()

    // 设置策略：单笔限额 1 ETH
    policy := &Policy{
        ID:           "policy-1",
        WalletID:     "wallet-1",
        SingleTxLimit: big.NewInt(1e18), // 1 ETH
    }
    err := engine.SetPolicy(ctx, policy)
    require.NoError(t, err)

    // 测试：0.5 ETH 应该通过
    req := &SignRequest{
        WalletID:  "wallet-1",
        To:        "0x1234567890123456789012345678901234567890",
        Value:     big.NewInt(5e17), // 0.5 ETH
        Timestamp: time.Now(),
    }
    err = engine.Check(ctx, req)
    assert.NoError(t, err)

    // 测试：2 ETH 应该失败
    req.Value = big.NewInt(2e18)
    err = engine.Check(ctx, req)
    assert.Error(t, err)
    assert.Equal(t, ErrExceedsSingleTxLimit, err)
}

func TestCheck_Whitelist(t *testing.T) {
    engine := setupTestEngine(t)
    ctx := context.Background()

    // 设置策略：只有 Uniswap 和 Curve 在白名单中
    policy := &Policy{
        ID:        "policy-1",
        WalletID:  "wallet-1",
        Whitelist: []string{
            "0xUniswap...",
            "0xCurve...",
        },
    }
    err := engine.SetPolicy(ctx, policy)
    require.NoError(t, err)

    // 测试：白名单地址应该通过
    req := &SignRequest{
        WalletID:  "wallet-1",
        To:        "0xuniswap...",
        Value:     big.NewInt(1e18),
        Timestamp: time.Now(),
    }
    err = engine.Check(ctx, req)
    assert.NoError(t, err)

    // 测试：非白名单地址应该失败
    req.To = "0xbad..."
    err = engine.Check(ctx, req)
    assert.Error(t, err)
    assert.Equal(t, ErrAddressNotWhitelisted, err)
}

func TestCheck_DailyLimit(t *testing.T) {
    engine := setupTestEngine(t)
    ctx := context.Background()

    // 设置策略：每日限额 10 ETH
    policy := &Policy{
        ID:         "policy-1",
        WalletID:   "wallet-1",
        DailyLimit: big.NewInt(10e18),
    }
    err := engine.SetPolicy(ctx, policy)
    require.NoError(t, err)

    // 模拟今日已使用 8 ETH
    today := time.Now()
    engine.storage.IncrementUsage(ctx, "wallet-1", today, big.NewInt(8e18))

    // 测试：1 ETH 应该通过
    req := &SignRequest{
        WalletID:  "wallet-1",
        To:        "0x1234...",
        Value:     big.NewInt(1e18),
        Timestamp: now,
    }
    err = engine.Check(ctx, req)
    assert.NoError(t, err)

    // 测试：3 ETH 应该失败（8 + 3 = 11 > 10）
    req.Value = big.NewInt(3e18)
    err = engine.Check(ctx, req)
    assert.Error(t, err)
    assert.Equal(t, ErrExceedsDailyLimit, err)
}

func TestCheck_TimeWindow(t *testing.T) {
    engine := setupTestEngine(t)
    ctx := context.Background()

    // 设置策略：只在工作时间允许
    startTime := time.Date(2026, 2, 21, 9, 0, 0, 0, time.UTC)
    endTime := time.Date(2026, 2, 21, 18, 0, 0, 0, time.UTC)

    policy := &Policy{
        ID:        "policy-1",
        WalletID:  "wallet-1",
        StartTime: &startTime,
        EndTime:   &endTime,
    }
    err := engine.SetPolicy(ctx, policy)
    require.NoError(t, err)

    // 测试：在工作时间内应该通过
    req := &SignRequest{
        WalletID:  "wallet-1",
        To:        "0x1234...",
        Value:     big.NewInt(1e18),
        Timestamp: time.Date(2026, 2, 21, 12, 0, 0, 0, time.UTC),
    }
    err = engine.Check(ctx, req)
    assert.NoError(t, err)

    // 测试：在工作时间外应该失败
    req.Timestamp = time.Date(2026, 2, 21, 20, 0, 0, 0, time.UTC)
    err = engine.Check(ctx, req)
    assert.Error(t, err)
    assert.Equal(t, ErrOutsideTimeWindow, err)
}
```

---

## API 集成

### 设置策略

```http
PUT /api/v1/wallet/:address/policy
Authorization: Bearer <API_KEY>
Content-Type: application/json

请求:
{
  "single_tx_limit": "1000000000000000000",  // 1 ETH (wei)
  "daily_limit": "10000000000000000000",     // 10 ETH
  "whitelist": [
    "0xUniswap...",
    "0xCurve..."
  ],
  "daily_tx_limit": 100
}
```

### 查询策略

```http
GET /api/v1/wallet/:address/policy
Authorization: Bearer <API_KEY>

响应:
{
  "success": true,
  "data": {
    "wallet_id": "wallet-1",
    "single_tx_limit": "1000000000000000000",
    "daily_limit": "10000000000000000000",
    "whitelist": ["0x...", "0x..."],
    "daily_tx_limit": 100
  }
}
```

### 查询每日使用

```http
GET /api/v1/wallet/:address/usage
Authorization: Bearer <API_KEY>

响应:
{
  "success": true,
  "data": {
    "date": "2026-02-21",
    "total_amount": "5000000000000000000",  // 已用 5 ETH
    "tx_count": 10
  }
}
```

---

## 完成标志

### 功能验证
- [ ] 所有策略检查正常工作
- [ ] 策略可以持久化存储
- [ ] 每日使用量正确累计
- [ ] 所有测试通过

### 代码质量
- [ ] `go test ./...` 通过
- [ ] `golangci-lint run` 通过
- [ ] 测试覆盖率 > 80%

### 集成验证
- [ ] API 可以设置策略
- [ ] 签名请求会被策略检查

---

## 下一步

完成后，可以开始 **Task 007: SDK 基础 + HTTP 客户端**
