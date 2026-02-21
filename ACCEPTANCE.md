# Task 011: API 文档 + 部署指南

**负责人**: Codex
**审核人**: Claude
**预计时间**: 2-3 小时
**依赖**: Task 010

---

## 功能要求

### 必须实现 (Must Have)
- [ ] API 文档（所有端点）
- [ ] SDK 使用文档
- [ ] Docker 部署指南
- [ ] 环境变量说明
- [ ] 快速开始指南
- [ ] 故障排查指南

### 建议实现 (Should Have)
- [ ] Postman 集合
- [ ] 架构图
- [ ] 性能优化指南
- [ ] 生产环境最佳实践

---

## 文档结构

```
docs/
├── api.md                     # API 文档
├── sdk.md                     # SDK 使用指南
├── deployment.md              # 部署指南
├── architecture.md            # 架构设计
├── security.md                # 安全最佳实践
├── troubleshooting.md         # 故障排查
└── openapi.json               # OpenAPI 规范
```

---

## API 文档 (api.md)

```markdown
# MPC Wallet API 文档

## 基础信息

- **Base URL**: `http://localhost:8080`
- **API Version**: v1
- **Content-Type**: `application/json`

## 认证

所有 API 请求需要在 Header 中包含 API Key:

\`\`\`http
Authorization: Bearer YOUR_API_KEY
\`\`\`

## 端点

### 1. 健康检查

检查服务健康状态。

\`\`\`http
GET /health
\`\`\`

**响应示例**:

\`\`\`json
{
  "status": "ok",
  "version": "0.1.0",
  "timestamp": "2026-02-21T10:00:00Z"
}
\`\`\`

### 2. 创建钱包

创建新的 MPC 钱包。

\`\`\`http
POST /api/v1/wallet/create
Content-Type: application/json
Authorization: Bearer YOUR_API_KEY

{
  "chain_id": "1"
}
\`\`\`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chain_id | number | 否 | 链 ID，默认 1 (Ethereum) |

**响应示例**:

\`\`\`json
{
  "success": true,
  "data": {
    "address": "0x1234567890123456789012345678901234567890",
    "public_key": "0x...",
    "shard1": "base64...",
    "shard2_id": "uuid-..."
  }
}
\`\`\`

### 3. 签名交易

签名交易或消息。

\`\`\`http
POST /api/v1/wallet/sign
Content-Type: application/json
Authorization: Bearer YOUR_API_KEY

{
  "address": "0x1234...7890",
  "message_hash": "0x...",
  "shard1": "base64..."
}
\`\`\`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| address | string | 是 | 钱包地址 |
| message_hash | string | 是 | 消息哈希（32 字节，十六进制） |
| shard1 | string | 是 | 密钥碎片 1（base64 编码） |

**响应示例**:

\`\`\`json
{
  "success": true,
  "data": {
    "signature": "0x...",
    "r": "0x...",
    "s": "0x...",
    "v": 28
  }
}
\`\`\`

**错误响应**:

\`\`\`json
{
  "success": false,
  "error": {
    "code": "EXCEEDS_SINGLE_TX_LIMIT",
    "message": "Transaction exceeds single transaction limit",
    "details": {
      "limit": "1000000000000000000",
      "value": "2000000000000000000"
    }
  }
}
\`\`\`

### 4. 查询钱包

获取钱包信息。

\`\`\`http
GET /api/v1/wallet/:address
Authorization: Bearer YOUR_API_KEY
\`\`\`

### 5. 设置策略

设置钱包策略。

\`\`\`http
PUT /api/v1/wallet/:address/policy
Content-Type: application/json
Authorization: Bearer YOUR_API_KEY

{
  "single_tx_limit": "1000000000000000000",
  "daily_limit": "10000000000000000000",
  "whitelist": ["0xUniswap...", "0xCurve..."],
  "daily_tx_limit": 100
}
\`\`\`

### 6. 查询策略

获取钱包策略。

\`\`\`http
GET /api/v1/wallet/:address/policy
Authorization: Bearer YOUR_API_KEY
\`\`\`

### 7. 查询使用情况

获取每日使用情况。

\`\`\`http
GET /api/v1/wallet/:address/usage
Authorization: Bearer YOUR_API_KEY
\`\`\`

## 错误代码

| 代码 | 说明 |
|------|------|
| UNAUTHORIZED | API Key 无效 |
| INVALID_REQUEST | 请求参数无效 |
| WALLET_NOT_FOUND | 钱包不存在 |
| INVALID_HASH | 消息哈希格式无效 |
| SHARD_NOT_FOUND | 密钥碎片未找到 |
| EXCEEDS_SINGLE_TX_LIMIT | 超过单笔交易限额 |
| EXCEEDS_DAILY_LIMIT | 超过每日交易限额 |
| EXCEEDS_DAILY_TX_LIMIT | 超过每日交易笔数限额 |
| ADDRESS_NOT_WHITELISTED | 目标地址不在白名单中 |
| OUTSIDE_TIME_WINDOW | 不在允许的时间窗口内 |
| SIGN_FAILED | 签名失败 |
```

---

## 部署指南 (deployment.md)

```markdown
# MPC Wallet 部署指南

## Docker 部署

### 快速开始

\`\`\`bash
# 1. 构建镜像
docker build -t agent-mpc-wallet:latest -f docker/Dockerfile .

# 2. 运行容器
docker run -d \
  -p 8080:8080 \
  -e MPC_API_KEYS=your-key-1,your-key-2 \
  -e SHARD_ENCRYPTION_KEY=$(openssl rand -base64 32) \
  -v $(pwd)/data:/app/data \
  agent-mpc-wallet:latest
\`\`\`

### Docker Compose

\`\`\`bash
docker-compose -f docker/docker-compose.yml up -d
\`\`\`

## 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| MPC_API_KEYS | 是 | - | 逗号分隔的 API Key 列表 |
| SHARD_ENCRYPTION_KEY | 是 | - | Shard 2 加密密钥（base64 编码，32 字节） |
| DB_PATH | 否 | ./data/mpc-wallet.db | SQLite 数据库路径 |
| SERVER_HOST | 否 | 0.0.0.0 | 服务器监听地址 |
| SERVER_PORT | 否 | 8080 | 服务器端口 |
| LOG_LEVEL | 否 | info | 日志级别 |

## 生产环境部署

### 使用 PostgreSQL

\`\`\`bash
docker run -d \\
  -e DB_TYPE=postgres \\
  -e DB_HOST=postgres.example.com \\
  -e DB_PORT=5432 \\
  -e DB_NAME=mpc_wallet \\
  -e DB_USER=mpc \\
  -e DB_PASSWORD=your-password \\
  -e SHARD_ENCRYPTION_KEY=$(openssl rand -base64 32) \\
  agent-mpc-wallet:latest
\`\`\`

### 使用 Kubernetes

\`\`\`yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mpc-wallet
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mpc-wallet
  template:
    metadata:
      labels:
        app: mpc-wallet
    spec:
      containers:
      - name: mpc-wallet
        image: agent-mpc-wallet:latest
        ports:
        - containerPort: 8080
        env:
        - name: MPC_API_KEYS
          valueFrom:
            secretKeyRef:
              name: mpc-secrets
              key: api-keys
        - name: SHARD_ENCRYPTION_KEY
          valueFrom:
            secretKeyRef:
              name: mpc-secrets
              key: encryption-key
        - name: DB_TYPE
          value: postgres
        - name: DB_HOST
          value: postgres-service
        volumeMounts:
        - name: data
          mountPath: /app/data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: mpc-data
---
apiVersion: v1
kind: Service
metadata:
  name: mpc-wallet
spec:
  selector:
    app: mpc-wallet
  ports:
  - port: 8080
    targetPort: 8080
  type: LoadBalancer
\`\`\`

## 健康检查

\`\`\`bash
# 健康检查端点
curl http://localhost:8080/health

# 预期响应
{"status":"ok","version":"0.1.0","timestamp":"..."}
\`\`\`

## 日志

\`\`\`bash
# 查看日志
docker logs mpc-wallet -f

# 或者使用 docker-compose
docker-compose logs -f mpc-wallet
\`\`\`

## 备份

### 数据库备份

\`\`\`bash
# SQLite
cp data/mpc-wallet.db backup/mpc-wallet-$(date +%Y%m%d).db

# PostgreSQL
pg_dump -h localhost -U mpc mpc_wallet > backup/mpc-$(date +%Y%m%d).sql
\`\`\`

### 恢复

\`\`\`bash
# SQLite
cp backup/mpc-wallet-20260221.db data/mpc-wallet.db

# PostgreSQL
psql -h localhost -U mpc mpc_wallet < backup/mpc-20260221.sql
\`\`\`
```

---

## 快速开始指南 (README.md 更新)

```markdown
# Agent MPC Wallet

> 开源的 AI Agent MPC 钱包 SDK

[![Go Tests](https://github.com/YOUR_USERNAME/agent-mpc-wallet/actions/workflows/go-test.yml/badge.svg)]
[![TS Tests](https://github.com/YOUR_USERNAME/agent-mpc-wallet/actions/workflows/ts-test.yml/badge.svg)]
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)]

## 特性

- 🔐 2-of-2 TSS 阈值签名
- 🚀 5 分钟快速集成
- 📡 HTTP/REST API
- 🔌 TypeScript SDK
- 🐳 Docker 一键部署
- 🛡️ 内置策略引擎

## 快速开始

### 1. 启动服务

\`\`\`bash
docker run -d \\
  -p 8080:8080 \\
  -e MPC_API_KEYS=your-api-key \\
  -e SHARD_ENCRYPTION_KEY=$(openssl rand -base64 32) \\
  ghcr.io/YOUR_USERNAME/agent-mpc-wallet:latest
\`\`\`

### 2. 安装 SDK

\`\`\`bash
npm install @agent-mpc-wallet/sdk
\`\`\`

### 3. 使用

\`\`\`typescript
import { MPCWallet } from '@agent-mpc-wallet/sdk';

const wallet = new MPCWallet({
  client: {
    baseURL: 'http://localhost:8080',
    apiKey: 'your-api-key',
  },
});

// 创建钱包
const address = await wallet.create();
console.log('Wallet address:', address);

// 签名消息
const signature = await wallet.signMessage('Hello, World!');
console.log('Signature:', signature);
\`\`\`

## 文档

- [API 文档](docs/api.md)
- [SDK 指南](docs/sdk.md)
- [部署指南](docs/deployment.md)
- [架构设计](docs/architecture.md)

## 示例

- [基础示例](examples/basic/README.md)
- [ElizaOS 集成](examples/with-eliza/README.md)
- [GOAT SDK 集成](examples/with-goat/README.md)

## 许可证

MIT License - see [LICENSE](LICENSE) for details.
```

---

## 架构文档 (architecture.md)

```markdown
# MPC Wallet 架构设计

## 系统架构

\`\`\`
┌─────────────────────────────────────────────────────────────┐
│  Agent 应用层                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  ElizaOS    │  │  GOAT SDK   │  │  Custom App │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│  SDK 层                                                     │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  MPCWallet (TypeScript)                               ││
│  │  - Shard 1 管理                                        ││
│  │  - 策略检查                                            ││
│  │  - HTTP 客户端                                         ││
│  └─────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│  API 层 (HTTP/REST)                                        │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  POST /api/v1/wallet/create                           ││
│  │  POST /api/v1/wallet/sign                             ││
│  │  PUT  /api/v1/wallet/:address/policy                  ││
│  └─────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│  服务层                                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │
│  │   TSS    │  │  Policy  │  │ Storage  │                 │
│  │   KeyGen │  │  Engine  │  │ SQLite   │                 │
│  │  Signing │  │          │  │ + AES    │                 │
│  └──────────┘  └──────────┘  └──────────┘                 │
├─────────────────────────────────────────────────────────────┤
│  数据层                                                    │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Shard 2 (AES-256-GCM 加密存储)                         ││
│  │  Wallet Info                                           ││
│  │  Policy Data                                           ││
│  │  Daily Usage                                           ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
\`\`\`

## TSS 流程

### 密钥生成

\`\`\`
1. Agent 生成临时密钥对
2. 发送 KeyGen 请求到 MPC 服务
3. MPC 服务创建 Party ID
4. 双方协同执行 GG18 协议
5. 输出:
   - 公钥 (计算地址)
   - Shard 1 (返回给 Agent)
   - Shard 2 (加密存储在服务端)
\`\`\`

### 签名流程

\`\`\`
1. Agent 发送签名请求 (message_hash + shard1)
2. MPC 服务:
   - 验证 shard1
   - 加载 shard2
   - 检查策略
   - 协同签名
3. 返回签名 (r, s, v)
\`\`\`

## 安全设计

### 密钥分片

| 分片 | 位置 | 保护措施 |
|------|------|---------|
| Shard 1 | Agent 环境 | 环境变量/K8s Secret |
| Shard 2 | MPC 服务 | AES-256-GCM + HSM |

### 策略引擎

- 单笔限额
- 每日限额
- 地址白名单
- 时间窗口
- 每日笔数限制

## 性能指标

| 操作 | 目标时间 |
|------|---------|
| 密钥生成 | < 10s (冷启动), < 3s (预计算) |
| 签名 | < 3s |
| API 响应 | < 100ms (不含签名) |
```

---

## OpenAPI 规范

```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "Agent MPC Wallet API",
    "version": "1.0.0",
    "description": "MPC Wallet API for AI Agents"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Development server"
    }
  ],
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check",
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/HealthResponse"
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/wallet/create": {
      "post": {
        "summary": "Create wallet",
        "security": [{"BearerAuth": []}],
        "requestBody": {
          "required": false,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/CreateWalletRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Wallet created",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/CreateWalletResponse"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "API Key"
      }
    },
    "schemas": {
      "HealthResponse": {
        "type": "object",
        "properties": {
          "status": {"type": "string"},
          "version": {"type": "string"},
          "timestamp": {"type": "string", "format": "date-time"}
        }
      }
    }
  }
}
```

---

## Postman 集合

```json
{
  "info": {
    "name": "MPC Wallet API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "http://localhost:8080/health"
        }
      }
    },
    {
      "name": "Create Wallet",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Authorization",
            "value": "Bearer {{apiKey}}"
          }
        ],
        "url": {
          "raw": "http://localhost:8080/api/v1/wallet/create"
        }
      }
    }
  ],
  "variable": [
    {
      "key": "apiKey",
      "value": "test-api-key"
    }
  ]
}
```

---

## 完成标志

### 文档完整性
- [ ] 所有 API 端点已文档化
- [ ] 所有 SDK 方法有说明
- [ ] 部署指南完整
- [ ] 有故障排查部分

### 示例完整性
- [ ] 快速开始指南
- [ ] 代码示例可运行
- [ ] 有输出说明

### 发布准备
- [ ] README.md 更新
- [ ] CHANGELOG.md 创建
- [ ] LICENSE 文件存在

---

## 项目完成检查清单

### Phase 1-4 全部完成
- [ ] Task 001: 项目骨架
- [ ] Task 002: TSS KeyGen
- [ ] Task 003: TSS Signing
- [ ] Task 004: HTTP Server
- [ ] Task 005: Shard 2 存储
- [ ] Task 006: 策略引擎
- [ ] Task 007: SDK 客户端
- [ ] Task 008: MPCWallet 类
- [ ] Task 009: 示例代码
- [ ] Task 010: E2E 测试
- [ ] Task 011: 文档

### v0.1.0 发布准备
- [ ] 所有测试通过
- [ ] 文档完整
- [ ] Docker 镜像可构建
- [ ] 测试网验证成功
- [ ] 安全审查完成

---

## 项目完成！

当所有任务完成后，即可发布 v0.1.0 版本。
