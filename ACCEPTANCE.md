# Task 005: Shard 2 存储层

**负责人**: Codex
**审核人**: Claude
**预计时间**: 3-4 小时
**依赖**: Task 002

---

## 功能要求

### 必须实现 (Must Have)
- [ ] SQLite 存储实现
- [ ] AES-256-GCM 加密存储 Shard 2
- [ ] CRUD 操作 (Create, Read, Update, Delete)
- [ ] 数据库迁移脚本
- [ ] 完整的单元测试

### 建议实现 (Should Have)
- [ ] PostgreSQL 支持（接口抽象）
- [ ] 连接池管理
- [ ] 备份机制

---

## 接口定义

```go
// server/internal/storage/storage.go
package storage

import (
    "context"
    "time"
)

// ShardStorage 密钥分片存储接口
type ShardStorage interface {
    // Store 存储 Shard 2（加密）
    Store(ctx context.Context, id string, shard2 []byte) error

    // Load 加载 Shard 2（解密）
    Load(ctx context.Context, id string) ([]byte, error)

    // Exists 检查是否存在
    Exists(ctx context.Context, id string) (bool, error)

    // Delete 删除 Shard 2
    Delete(ctx context.Context, id string) error

    // List 列出所有钱包地址
    List(ctx context.Context) ([]string, error)
}

// WalletInfo 钱包信息
type WalletInfo struct {
    ID            string    `json:"id"`             // 钱包 ID
    Address       string    `json:"address"`        // 以太坊地址
    PublicKey     string    `json:"public_key"`     // 公钥
    Shard2ID      string    `json:"shard2_id"`      // Shard 2 的存储 ID
    CreatedAt     time.Time `json:"created_at"`     // 创建时间
    UpdatedAt     time.Time `json:"updated_at"`     // 更新时间
}

// WalletStorage 钱包存储接口
type WalletStorage interface {
    // Create 创建钱包记录
    Create(ctx context.Context, info *WalletInfo) error

    // GetByAddress 根据地址获取钱包信息
    GetByAddress(ctx context.Context, address string) (*WalletInfo, error)

    // GetByID 根据 ID 获取钱包信息
    GetByID(ctx context.Context, id string) (*WalletInfo, error)

    // Update 更新钱包信息
    Update(ctx context.Context, info *WalletInfo) error

    // Delete 删除钱包
    Delete(ctx context.Context, id string) error
}
```

---

## 数据库 Schema

```sql
-- migrations/001_initial_schema.sql

CREATE TABLE wallets (
    id TEXT PRIMARY KEY,
    address TEXT UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    shard2_id TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX idx_wallets_address ON wallets(address);
CREATE INDEX idx_wallets_shard2_id ON wallets(shard2_id);

CREATE TABLE key_shards (
    id TEXT PRIMARY KEY,
    shard2_encrypted BLOB NOT NULL,
    nonce BLOB NOT NULL,  -- GCM nonce
    created_at INTEGER NOT NULL
);
```

---

## 加密要求

```go
// server/internal/storage/encrypt.go
package storage

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

// Encryptor 加密器接口
type Encryptor interface {
    // Encrypt 加密数据
    Encrypt(plaintext []byte) ([]byte, error)

    // Decrypt 解密数据
    Decrypt(ciphertext []byte) ([]byte, error)
}

// AES256GCMEncryptor AES-256-GCM 加密器
type AES256GCMEncryptor struct {
    key []byte // 32 字节密钥
}

// NewAES256GCMEncryptor 创建加密器
// 密钥从环境变量 SHARD_ENCRYPTION_KEY 读取（base64 编码的 32 字节）
func NewAES256GCMEncryptor(key []byte) (*AES256GCMEncryptor, error) {
    if len(key) != 32 {
        return nil, ErrInvalidKeySize
    }
    return &AES256GCMEncryptor{key: key}, nil
}

// Encrypt 加密（AES-256-GCM）
func (e *AES256GCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}

// Decrypt 解密
func (e *AES256GCMEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, ErrInvalidCiphertext
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

---

## SQLite 实现

```go
// server/internal/storage/sqlite.go
package storage

import (
    "context"
    "database/sql"
    "embed"
    "fmt"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// SQLiteStorage SQLite 实现
type SQLiteStorage struct {
    db        *sql.DB
    encryptor Encryptor
}

// NewSQLiteStorage 创建 SQLite 存储
func NewSQLiteStorage(dbPath string, encryptor Encryptor) (*SQLiteStorage, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // 启用外键约束
    db.Exec("PRAGMA foreign_keys = ON")

    storage := &SQLiteStorage{
        db:        db,
        encryptor: encryptor,
    }

    // 运行迁移
    if err := storage.migrate(); err != nil {
        return nil, err
    }

    return storage, nil
}

// migrate 运行数据库迁移
func (s *SQLiteStorage) migrate() error {
    // 实现迁移逻辑
    // ...
}
```

---

## 配置要求

```go
// server/internal/storage/config.go
package storage

type Config struct {
    Type     string `mapstructure:"type"`     // sqlite, postgres
    Path     string `mapstructure:"path"`     // 数据库路径
    Host     string `mapstructure:"host"`     // PostgreSQL host
    Port     int    `mapstructure:"port"`     // PostgreSQL port
    Database string `mapstructure:"database"` // Database name
    User     string `mapstructure:"user"`     // User
    Password string `mapstructure:"password"` // Password
}
```

---

## 环境变量

```bash
# 加密密钥（base64 编码的 32 字节）
SHARD_ENCRYPTION_KEY=<base64-encoded-32-bytes>

# 数据库路径
DB_PATH=./data/mpc-wallet.db

# 或使用 PostgreSQL
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mpc_wallet
DB_USER=postgres
DB_PASSWORD=postgres
```

---

## 测试用例

```go
// server/internal/storage/sqlite_test.go
package storage

import (
    "context"
    "os"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *SQLiteStorage {
    // 使用内存数据库
    key := make([]byte, 32)
    encryptor, err := NewAES256GCMEncryptor(key)
    require.NoError(t, err)

    storage, err := NewSQLiteStorage(":memory:", encryptor)
    require.NoError(t, err)

    return storage
}

func TestStoreAndLoad(t *testing.T) {
    storage := setupTestDB(t)
    ctx := context.Background()

    shard2 := []byte("test-shard-data")
    id := "test-id"

    // 存储
    err := storage.Store(ctx, id, shard2)
    require.NoError(t, err)

    // 加载
    loaded, err := storage.Load(ctx, id)
    require.NoError(t, err)
    assert.Equal(t, shard2, loaded)
}

func TestEncryption(t *testing.T) {
    key := make([]byte, 32)
    encryptor, err := NewAES256GCMEncryptor(key)
    require.NoError(t, err)

    plaintext := []byte("sensitive-shard-data")

    // 加密
    ciphertext, err := encryptor.Encrypt(plaintext)
    require.NoError(t, err)

    // 密文不同
    assert.NotEqual(t, plaintext, ciphertext)

    // 解密
    decrypted, err := encryptor.Decrypt(ciphertext)
    require.NoError(t, err)
    assert.Equal(t, plaintext, decrypted)
}

func TestWalletCRUD(t *testing.T) {
    storage := setupTestDB(t)
    ctx := context.Background()

    info := &WalletInfo{
        ID:        "wallet-1",
        Address:   "0x1234567890123456789012345678901234567890",
        PublicKey: "0x...",
        Shard2ID:  "shard-2",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Create
    err := storage.Create(ctx, info)
    require.NoError(t, err)

    // GetByAddress
    loaded, err := storage.GetByAddress(ctx, info.Address)
    require.NoError(t, err)
    assert.Equal(t, info.Address, loaded.Address)

    // GetByID
    loaded, err = storage.GetByID(ctx, info.ID)
    require.NoError(t, err)
    assert.Equal(t, info.ID, loaded.ID)

    // Update
    loaded.UpdatedAt = time.Now()
    err = storage.Update(ctx, loaded)
    require.NoError(t, err)

    // Delete
    err = storage.Delete(ctx, info.ID)
    require.NoError(t, err)

    // 验证删除
    _, err = storage.GetByID(ctx, info.ID)
    assert.Error(t, err)
}
```

---

## 完成标志

### 功能验证
- [ ] Shard 2 加密后存储
- [ ] 可以正确读取和解密
- [ ] CRUD 操作正常
- [ ] 数据库迁移脚本可执行

### 代码质量
- [ ] `go test ./...` 通过
- [ ] `golangci-lint run` 通过
- [ ] 测试覆盖率 > 80%

### 安全验证
- [ ] 密钥从环境变量读取
- [ ] 硬编码密钥测试失败
- [ ] 解密错误数据返回错误

---

## 下一步

完成后，可以开始 **Task 006: 策略引擎**
