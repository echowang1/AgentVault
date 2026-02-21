# Task 001 给 Codex 的任务说明

**发给 Codex 的完整说明**

---

## 一、项目背景

### 项目是什么

**AgentVault** - 一个面向 AI Agent 的开源 MPC 钱包 SDK。

**核心功能**:
- 2-of-2 TSS（阈值签名）
- AI Agent 可以独立控制资产
- 策略引擎（限额、白名单等安全控制）

**为什么做这个**：
- AI Agent 需要拥有独立的资产控制权
- 现有钱包方案都是为人类设计的（社交登录、UI 交互）
- Agent 需要的是 API-first 的解决方案

### 你之前的经验

- 你之前做过面向 C 端用户的 MPC 钱包（OpenBlock）
- 技术栈类似，但用户群体完全不同（现在是 Agent 开发者）
- 可以参考 OpenBlock 的安全设计，但要重新设计交互模式

---

## 二、技术架构

```
┌─────────────────────────────────────────────────────────────┐
│  2-of-2 TSS 架构                                           │
│                                                             │
│  Shard 1 (Agent 环境)           Shard 2 (MPC 服务)          │
│  ├─ 环境变量                  ├─ 你的服务托管               │
│  ├─ K8s Secret                ├─ HSM/TEE 可选               │
│  └─ 内存加载                  └─ 开发者自托管               │
│                                                             │
│  签名流程: Agent 请求 ──▶ MPC 服务 ──▶ 协同签名 ──▶ 返回   │
└─────────────────────────────────────────────────────────────┘
```

### 技术选型

| 组件 | 技术 | 说明 |
|------|------|------|
| **TSS 库** | bnb-chain/tss-lib | MIT 许可，GG18 协议 |
| **后端** | Go 1.21+ | tss-lib 原生语言 |
| **SDK** | TypeScript | Agent 生态主流 |
| **存储** | SQLite → PostgreSQL | 先轻量，后扩展 |
| **API** | HTTP/REST + JSON | 简单易用 |

---

## 三、当前任务：Task 001 项目骨架搭建

### 任务目标

创建项目的基础骨架，让后续开发可以顺利进行。

**这不是**写核心功能，只是搭架子。

### 你需要做的事情

#### 1. 完善 Go Server 配置

**文件**: `server/cmd/server/main.go`

```go
package main

import (
    "fmt"
    "log"
    "os"
)

func main() {
    // TODO: 后续会在这里集成 gin 路由和 TSS 服务
    // 现在只需要能跑起来

    fmt.Println("MPC Wallet Server v0.1.0")
    fmt.Println("Server starting on port 8080...")

    // 简单的 HTTP 服务器占位
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(200)
        w.Write([]byte(`{"status":"ok","version":"0.1.0"}`))
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

#### 2. 完善配置文件

**文件**: `server/internal/config/config.go`

```go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    ServerHost string
    ServerPort int
    APIKeys    map[string]bool
}

func Load() (*Config, error) {
    cfg := &Config{
        ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
        ServerPort: getEnvInt("SERVER_PORT", 8080),
    }

    // 解析 API Keys（逗号分隔）
    apiKeyStr := getEnv("MPC_API_KEYS", "")
    if apiKeyStr != "" {
        cfg.APIKeys = make(map[string]bool)
        for _, key := range split(apiKeyStr, ",") {
            cfg.APIKeys[trim(key)] = true
        }
    }

    return cfg, nil
}

func getEnv(key, defaultValue string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if val := os.Getenv(key); val != "" {
        if intVal, err := strconv.Atoi(val); err == nil {
            return intVal
        }
    }
    return defaultValue
}
```

#### 3. 创建 Dockerfile

**文件**: `docker/Dockerfile`

```dockerfile
# Multi-stage build for Go server
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY server/go.mod server/go.sum* ./
RUN go mod download

COPY server/ ./server/
RUN CGO_ENABLED=0 go build -o server ./server/cmd/server

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/docker/entrypoint.sh .

EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
```

**文件**: `docker/entrypoint.sh`

```bash
#!/bin/sh
set -e

echo "Starting MPC Wallet Server..."
exec /app/server
```

#### 4. 添加 SDK 基础导出

**文件**: `sdk/src/index.ts`

```typescript
// TODO: 后续会实现完整的 MPC Wallet 类
// 现在只需要基础导出

export const PACKAGE_NAME = '@agent-vault/sdk';
export const VERSION = '0.1.0-alpha';

// 占位接口
export interface WalletConfig {
  baseURL: string;
  apiKey: string;
}

export interface SignRequest {
  address: string;
  messageHash: string;
  shard1: string;
}

export interface SignResponse {
  signature: string;
  r: string;
  s: string;
  v: number;
}
```

#### 5. 更新 docker-compose.yml

**文件**: `docker/docker-compose.yml`

```yaml
version: '3.8'

services:
  mpc-wallet:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
      - MPC_API_KEYS=test-api-key
    volumes:
      - ./data:/app/data
```

#### 6. 添加基础测试

**文件**: `server/cmd/server/main_test.go`

```go
package main

import "testing"

func TestMainRuns(t *testing.T) {
    // 占位测试，确保可以编译
    if true {
        return
    }
    // 实际启动测试在后续任务
}
```

**文件**: `sdk/index.test.ts`

```typescript
import { describe, it } from 'vitest';
import { PACKAGE_NAME, VERSION } from './index';

describe('SDK', () => {
  it('should export package name', () => {
    expect(PACKAGE_NAME).toBe('@agent-vault/sdk');
  });

  it('should export version', () => {
    expect(VERSION).toBe('0.1.0-alpha');
  });
});
```

---

## 四、验收标准

完成任务后，以下应该都能通过：

### 1. 构建测试
```bash
cd /Users/echo/agent-vault-task-001
make build
# 应该成功编译 server 和 sdk
```

### 2. Docker 测试
```bash
make docker-build
# 应该成功构建镜像
```

### 3. 运行测试
```bash
make test
# 应该有测试通过（虽然现在是占位测试）
```

### 4. 服务启动测试
```bash
cd server && go run cmd/server/main.go
# 访问 http://localhost:8080/health 应该返回 JSON
```

---

## 五、重要注意事项

### ❌ 不要做的事情

1. **不要**实现 TSS 核心逻辑（那是 Task 002/003）
2. **不要**实现完整的 HTTP API（那是 Task 004）
3. **不要**修改已定的目录结构
4. **不要**添加额外的依赖（除非必要）

### ✅ 要做的事情

1. 确保 `make build` 能成功
2. 确保 Docker 镜像能构建
3. 确保所有占位文件存在
4. 确保测试框架配置正确

### 📝 代码风格

- Go: 使用 `gofmt` 格式化
- TypeScript: 使用 `prettier` 格式化
- 注释：复杂逻辑需要注释，简单代码不需要

---

## 六、提交规范

提交信息格式：
```
[task-001] feat: 项目骨架搭建

- 添加 Go server 基础结构
- 添加 TypeScript SDK 基础结构
- 配置 Docker 和 docker-compose
- 配置 GitHub Actions CI
```

---

## 七、遇到问题怎么办

1. **不确定某个文件怎么写** → 先写一个最小可运行的版本，在提交信息中标注 "TODO: 后续完善"
2. **发现目录结构不够** → 可以在 PR 中说明，等审核时再调整
3. **测试跑不通** → 先保证代码可以编译，测试框架配置正确

---

## 八、完成后

1. 运行 `make test` 确保通过
2. 运行 `make build` 确保编译成功
3. 提交代码
4. 告诉我进行验收

---

## 附录：参考文档位置

- **项目规划**: `/Users/echo/claudesidian/01_Projects/specs/005-agent-vault/12-development-roadmap.md`
- **协作规范**: `/Users/echo/claudesidian/01_Projects/specs/005-agent-vault/13-acceptance-standards-template.md`
- **本任务验收标准**: `/Users/echo/claudesidian/01_Projects/specs/005-agent-vault/tasks/001-init-skeleton/ACCEPTANCE.md`
