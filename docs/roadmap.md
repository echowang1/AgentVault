# AgentVault Roadmap

## 版本策略

- **v0.1.x** - 稳定性修复和小改进
- **v0.2.0** - 稳定性硬化
- **v0.3.0** - 可观测性
- **v0.4.0** - 生产增强
- **v1.0.0** - 企业特性

---

## 优先级改进计划

### 1. 稳定性硬化 (Stability)

#### 1.1 拆分 E2E 测试为 smoke/long 两档

**问题**: 当前所有 E2E 测试都在 PR 中运行，最慢的性能测试容易超时。

**方案**:
- `tests/smoke.test.ts` - 快速冒烟测试 (< 30s)
- `tests/long.test.ts` - 完整性能测试 (可超时)
- PR 只跑 smoke，nightly 跑 long

**文件**:
- `tests/smoke.test.ts` (新建)
- `tests/long.test.ts` (新建)
- `.github/workflows/smoke.yml` (新建)
- `.github/workflows/nightly.yml` (新建)

---

### 2. SDK/API 一致性修复

#### 2.1 Health 响应语义统一

**问题**: `health` endpoint 和 SDK `healthCheck()` 解析不一致。

**当前状态**:
```typescript
// API 返回
{ "status": "ok" }

// SDK 期望
{ success: true; data: { status: string } }
```

**方案**: 统一为一种格式，建议简化为直出：
```json
{ "status": "ok", "version": "0.1.0" }
```

**文件**:
- `server/handlers/health.go`
- `sdk/src/client.ts`

---

### 3. 错误码标准化

#### 3.1 细化策略错误码

**问题**: 策略相关错误都归为 `INVALID_REQUEST`，无法区分具体原因。

**新增错误码**:
| 错误码 | 说明 |
|--------|------|
| `POLICY_SINGLE_TX_LIMIT_EXCEEDED` | 单笔交易超限 |
| `POLICY_DAILY_LIMIT_EXCEEDED` | 每日总额超限 |
| `POLICY_WHITELIST_DENIED` | 地址不在白名单 |
| `POLICY_DAILY_TX_COUNT_EXCEEDED` | 每日笔数超限 |
| `POLICY_TIME_WINDOW_VIOLATION` | 时间窗口限制 |

**文件**:
- `server/errors/errors.go`
- `docs/openapi.json`
- `sdk/src/errors.ts`

---

### 4. 安全基线升级

#### 4.1 默认速率限制

**方案**: 添加中间件，默认限制：
- 每个 API key: 100 req/min
- 每个 IP: 200 req/min

**文件**:
- `server/middleware/rate_limit.go` (新建)

#### 4.2 审计日志

**字段**:
- `wallet_id` - 钱包地址
- `request_id` - 请求追踪 ID
- `action` - 操作类型 (keygen/sign/policy_update)
- `policy_hit` - 命中的策略规则
- `timestamp` - 时间戳

**文件**:
- `server/audit/logger.go` (新建)

#### 4.3 敏感字段脱敏

**脱敏规则**:
- API key: `test-api***`
- Address: `0x742d***beb0`
- Shard: `[REDACTED]`

**文件**:
- `server/middleware/logging.go`

#### 4.4 密钥轮换文档

**文件**:
- `docs/security.md` (更新)
- `docs/deployment.md` (更新)

---

### 5. 存储可靠性

#### 5.1 SQLite 连接池和 PRAGMA 配置

**配置**:
```go
db.SetMaxOpenConns(1)    // SQLite 写操作必须单连接
db.SetMaxIdleConns(1)
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA synchronous=NORMAL")
db.Exec("PRAGMA busy_timeout=5000")
```

**文件**:
- `server/internal/storage/sqlite.go`

#### 5.2 启动自检

**检查项**:
- 迁移完整性
- 文件权限
- 可写性测试

**文件**:
- `server/internal/storage/health.go` (新建)

---

### 6. 覆盖率与回归护栏

#### 6.1 关键路径回归测试矩阵

| 模块 | 测试场景 |
|------|----------|
| TSS | 边界值、并发、超时、断连恢复 |
| Policy | 各策略边界值、组合策略 |
| API | 无效请求、超长请求、并发 |

**文件**:
- `server/internal/tss/keygen_test.go` (扩展)
- `server/internal/policy/engine_test.go` (扩展)
- `server/internal/handlers/handlers_test.go` (扩展)

#### 6.2 CI 覆盖率门槛

**配置**:
```yaml
- .github/workflows/test.yml
  coverage:
    - statement: 70%
    - branch: 65%
```

---

### 7. 可观测性

#### 7.1 Prometheus 指标

| 指标 | 类型 | 标签 |
|------|------|------|
| `keygen_duration_seconds` | Histogram | success |
| `sign_duration_seconds` | Histogram | success, policy_hit |
| `policy_reject_total` | Counter | reason |
| `storage_errors_total` | Counter | operation |
| `active_wallets` | Gauge | - |

**文件**:
- `server/metrics/prometheus.go` (新建)
- `server/cmd/server/main.go` (集成)

#### 7.2 结构化日志

**方案**: 使用 `zap` 或 `zerolog`

**文件**:
- `server/log/logger.go` (新建)
- `server/...` (全部替换 log.Println)

---

### 8. 发布工程化

#### 8.1 CHANGELOG.md

**格式**: 遵循 [Keep a Changelog](https://keepachangelog.com/)

**文件**:
- `CHANGELOG.md` (新建)

#### 8.2 语义化版本策略

**规则**:
- MAJOR: 不兼容的 API 变更
- MINOR: 向后兼容的功能新增
- PATCH: 向后兼容的问题修复

**文件**:
- `docs/versioning.md` (新建)

#### 8.3 Release Workflow

**触发**: Git tag 推送

**动作**:
1. 构建 SDK npm 包
2. 构建 Docker image
3. 生成 GitHub Release
4. 自动更新 CHANGELOG

**文件**:
- `.github/workflows/release.yml` (新建)

---

## 完成状态

| 任务 | 状态 | 负责人 |
|------|------|--------|
| 稳定性硬化 - E2E 拆分 | ⏳ 待开始 | - |
| SDK/API 一致性修复 | ⏳ 待开始 | - |
| 错误码标准化 | ⏳ 待开始 | - |
| 安全基线升级 | ⏳ 待开始 | - |
| 存储可靠性 | ⏳ 待开始 | - |
| 覆盖率与回归护栏 | ⏳ 待开始 | - |
| 可观测性 | ⏳ 待开始 | - |
| 发布工程化 | ⏳ 待开始 | - |

---

**最后更新**: 2026-02-21
