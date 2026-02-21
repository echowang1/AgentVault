# 与 Codex 协作指南

**如何高效、准确地分配和验收任务**

---

## 一、任务发起模板

每次给 Codex 任务时，使用以下结构：

### 1.1 基础模板

```
【任务目标】
简要描述要做什么（1-2句话）

【上下文】
- 项目：AgentVault（AI Agent MPC 钱包）
- 当前分支：task/001-init-skeleton
- 工作目录：/Users/echo/AgentVault-task-001

【任务说明】
指向具体的验收标准文件，让 Codex 自己阅读

【注意事项】
1. 先阅读 CODEX_INSTRUCTIONS.md 了解项目背景
2. 只做当前任务要求的内容，不要过度设计
3. 遇到不确定的地方，先问我或标注 TODO

【完成后】
1. 运行 make test 确保通过
2. 告诉我已完成，等待验收
```

### 1.2 Task 001 具体示例

```
【任务目标】
完成 AgentVault 项目骨架搭建，让后续开发可以顺利进行

【上下文】
- 项目：AgentVault（AI Agent MPC 钱包 SDK）
- 当前分支：task/001-init-skeleton
- 工作目录：/Users/echo/AgentVault-task-001
- 这是一个全新项目，从零开始搭建

【任务说明】
请阅读以下文件了解详细要求：
1. /Users/echo/AgentVault-task-001/CODEX_INSTRUCTIONS.md（项目背景和任务说明）
2. /Users/echo/claudesidian/01_Projects/specs/005-agent-vault/tasks/001-init-skeleton/ACCEPTANCE.md（验收标准）

【关键要求】
- 这只是搭架子，不实现核心 TSS 功能（那是 Task 002/003）
- 确保 make build 能成功编译
- 确保 Docker 镜像能构建
- 创建完整的目录结构和配置文件

【注意事项】
1. 先阅读 CODEX_INSTRUCTIONS.md 了解项目背景
2. 只做当前任务要求的内容，不要过度设计
3. 代码风格：Go 用 gofmt，TS 用 prettier
4. 遇到不确定的地方，先问我或标注 TODO

【完成后】
1. cd /Users/echo/AgentVault-task-001
2. make test（确保通过）
3. make build（确保编译成功）
4. git status 查看修改的文件
5. 告诉我已完成
```

---

## 二、上下文提供清单

### 2.1 必须提供的上下文

| 项目 | 说明 |
|------|------|
| **项目名称** | AgentVault |
| **项目简介** | 面向 AI Agent 的 MPC 钱包 SDK |
| **当前任务** | Task 001: 项目骨架搭建 |
| **验收标准** | 具体文件路径 |
| **工作目录** | 绝对路径 |
| **分支名称** | task/001-init-skeleton |

### 2.2 项目架构概览（每次都提供）

```
AgentVault 技术架构：

后端: Go 1.21 + tss-lib + Gin
SDK: TypeScript + Vitest
存储: SQLite → PostgreSQL
API: HTTP/REST + JSON

目录结构：
/server/cmd/server    # Go 入口
/server/internal/tss    # TSS 核心封装
/server/internal/api    # HTTP API
/server/internal/storage # Shard 2 存储
/server/internal/policy  # 策略引擎
/sdk/src               # TypeScript SDK
```

---

## 三、监控进度

### 3.1 实时监控命令

```bash
# 在另一个终端监控
cd /Users/echo/AgentVault-task-001

# 查看最近的提交
watch -n 5 'git log --oneline -5'

# 查看修改的文件
watch -n 5 'git status --short'
```

### 3.2 定期检查点

| 阶段 | 检查内容 |
|------|---------|
| **开始前** | Codex 确认已阅读 CODEX_INSTRUCTIONS.md |
| **中期** | git status 查看是否有大范围修改 |
| **完成后** | 代码审查 + 运行测试 |

---

## 四、避免偏差的关键点

### 4.1 明确边界

```
DO ✓
- 只实现当前任务要求的功能
- 遇到不确定的地方标注 TODO
- 提前沟通需求不清晰的地方

DON'T ✗
- 提前实现后续任务的功能
- 过度设计或过度优化
- 修改已定的目录结构
- 添加额外的依赖
```

### 4.2 代码风格约定

```go
// Go: 直接返回错误，不要包装
if err != nil {
    return nil, err
}
```

```typescript
// TypeScript: 使用简洁的类型
type WalletConfig = {
  baseURL: string;
  apiKey: string;
};
```

### 4.3 提交规范

```
[task-001] type: description

示例：
[task-001] feat: 添加 Docker 配置
[task-001] fix: 修正 Makefile 路径
[task-001] chore: 更新 README
```

---

## 五、验收流程

### 5.1 我验收时检查的内容

1. **代码审查**
   - 是否符合验收标准
   - 是否有过度设计
   - 代码风格是否一致

2. **运行测试**
   ```bash
   cd /Users/echo/AgentVault-task-001
   make test
   make build
   ```

3. **检查文件**
   ```bash
   git status
   git diff --stat
   ```

### 5.2 通过后的处理

```bash
# 1. 我审查通过后，切换到 main 分支
cd /Users/echo/AgentVault
git checkout main

# 2. 合并任务分支
git merge task/001-init-skeleton

# 3. 推送
git push origin main
git push origin task/001-init-skeleton

# 4. 删除 worktree
git worktree remove ../AgentVault-task-001
git branch -d task/001-init-skeleton
```

---

## 六、常见问题处理

### 6.1 Codex 问"需要什么依赖？"

```
回答：
Go: github.com/bnb-chain/tss-lib, github.com/gin-gonic/gin, github.com/ethereum/go-ethereum
TS: axios, vitest, typescript, prettier, eslint

不要添加其他依赖，除非任务明确要求。
```

### 6.2 Codex 问"是否需要实现 X 功能？"

```
回答：
只实现当前任务 ACCEPTANCE.md 中明确列出的功能。
X 功能属于 Task XXX，后续会实现。

如果当前任务无法完成，先做最小可行版本，标注 TODO。
```

### 6.3 Codex 问"代码放在哪个文件？"

```
回答：
严格按照 CODEX_INSTRUCTIONS.md 中的文件结构。

如果文件不存在，创建它。
如果文件已存在，修改它（不要删除现有内容）。
```

---

## 七、快速命令参考

```bash
# === 目录导航 ===
cd /Users/echo/AgentVault-task-001    # 切换到 worktree
cd /Users/echo/AgentVault             # 切换到主仓库

# === Git 操作 ===
git status                            # 查看修改
git log --oneline -5                  # 查看最近提交
git diff                              # 查看具体修改

# === 测试和构建 ===
make test                             # 运行测试
make build                            # 编译项目
make docker-build                     # 构建 Docker

# === 同步代码 ===
git pull origin main                   # 同步主分支
git push origin task/001-init-skeleton  # 推送任务分支
```

---

## 八、第一次完整对话示例

---

### 你对 Codex 说：

```
【任务目标】
完成 AgentVault 项目骨架搭建

【上下文】
- 项目：AgentVault（AI Agent MPC 钱包 SDK）
- 当前分支：task/001-init-skeleton
- 工作目录：/Users/echo/AgentVault-task-001

【任务说明】
请先阅读以下文件了解完整要求：
1. CODEX_INSTRUCTIONS.md（项目背景和详细说明）
2. 参考 ACCEPTANCE.md 中的验收标准

【关键点】
- 这只是搭架子，不实现核心功能
- 确保 make build 和 make test 能通过
- 创建 Docker 配置

【完成后】
运行 make test，然后告诉我已完成
```

---

### Codex 可能的问题和你的回答

**Codex**: "我需要先了解项目架构"

**你**: "请阅读 CODEX_INSTRUCTIONS.md 文件，里面有完整的架构说明。"

**Codex**: "是否需要实现 TSS 签名功能？"

**你**: "不需要。那是 Task 002 和 Task 003 的工作。当前任务只搭架子。"

**Codex**: "应该用哪个 Go 框架？"

**你**: "gin-gonic/gin。这是在 CODEX_INSTRUCTIONS.md 中已确定的。"

**Codex**: "完成了，需要测试吗？"

**你**: "是的。请运行 make test，然后告诉我结果。"
