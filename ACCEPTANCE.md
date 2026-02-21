# Task 001: 项目骨架搭建

**负责人**: Codex
**审核人**: Claude
**预计时间**: 2-3 小时
**依赖**: 无

---

## 功能要求

### 必须实现 (Must Have)
- [ ] 创建 GitHub 仓库标准文件结构
- [ ] Go module 初始化（server/）
- [ ] TypeScript project 初始化（sdk/）
- [ ] Makefile 配置
- [ ] Docker 配置
- [ ] CI/CD 配置（GitHub Actions）
- [ ] 基础 README.md

### 建议实现 (Should Have)
- [ ] Pre-commit hooks (gofmt, eslint)
- [ ] .editorconfig
- [ ] 贡献者指南

---

## 目录结构

```
agent-mpc-wallet/
├── .github/
│   ├── workflows/
│   │   ├── go-test.yml
│   │   └── ts-test.yml
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   └── PULL_REQUEST_TEMPLATE.md
├── server/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go       # 最小可运行的 server
│   ├── internal/
│   │   └── config/
│   │       └── config.go
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── sdk/
│   ├── src/
│   │   └── index.ts          # 最小导出
│   ├── package.json
│   ├── tsconfig.json
│   └── tsconfig.build.json
├── tests/
│   └── placeholder.test.ts
├── .gitignore
├── .editorconfig
├── Makefile
├── docker-compose.yml
├── LICENSE (MIT)
└── README.md
```

---

## 配置要求

### Makefile 目标

```makefile
.PHONY: all build test clean docker-build docker-run

all: build

build:
	@echo "Building server..."
	cd server && go build -o ../bin/server ./cmd/server
	@echo "Building SDK..."
	cd sdk && npm run build

test:
	@echo "Running server tests..."
	cd server && go test ./...
	@echo "Running SDK tests..."
	cd sdk && npm test

clean:
	rm -rf bin/
	cd sdk && rm -rf dist/

docker-build:
	docker build -t agent-mpc-wallet:latest -f server/Dockerfile .

docker-run:
	docker run -p 8080:8080 agent-mpc-wallet:latest

lint-server:
	cd server && golangci-lint run

lint-sdk:
	cd sdk && npm run lint

format-server:
	cd server && gofmt -s -w .

format-sdk:
	cd sdk && npm run format
```

### Go 版本
- Go 1.21+
- 使用 `github.com/gin-gonic/gin`
- 使用 `github.com/stretchr/testify` 测试

### TypeScript 版本
- Node.js 18+
- TypeScript 5.3+
- Vitest 测试框架
- ESLint + Prettier

### Docker
- Multi-stage build
- 基于 `golang:1.21-alpine`
- 最终镜像 < 50MB

---

## 测试用例

### Go 测试
```go
// server/cmd/server/main.go_test.go
package main

import "testing"

func TestMainCanRun(t *testing.T) {
    // 最小验证：main 函数可以编译
    // 实际启动测试在集成测试中
}
```

### TypeScript 测试
```typescript
// sdk/index.test.ts
import { describe, it } from 'vitest';

describe('SDK', () => {
  it('should export something', () => {
    // 最小验证：模块可以加载
  });
});
```

---

## CI/CD 要求

### GitHub Actions - Go 测试

```yaml
# .github/workflows/go-test.yml
name: Go Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: cd server && go test ./...
      - run: cd server && golangci-lint run
```

### GitHub Actions - TypeScript 测试

```yaml
# .github/workflows/ts-test.yml
name: TypeScript Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: cd sdk && npm ci
      - run: cd sdk && npm test
      - run: cd sdk && npm run lint
```

---

## README.md 要求

```markdown
# Agent MPC Wallet

[![Go Tests](https://github.com/YOUR_USERNAME/agent-mpc-wallet/actions/workflows/go-test.yml/badge.svg)](https://github.com/YOUR_USERNAME/agent-mpc-wallet/actions/workflows/go-test.yml)
[![TS Tests](https://github.com/YOUR_USERNAME/agent-mpc-wallet/actions/workflows/ts-test.yml/badge.svg)](https://github.com/YOUR_USERNAME/agent-mpc-wallet/actions/workflows/ts-test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> 开源的 AI Agent MPC 钱包 SDK

## 特性

- 🔐 2-of-2 TSS 阈值签名
- 🚀 5 分钟快速集成
- 📡 HTTP/REST API
- 🔌 TypeScript SDK
- 🐳 Docker 一键部署

## 快速开始

### 服务端

\`\`\`bash
make build
make test
make docker-build
\`\`\`

### SDK

\`\`\`bash
cd sdk
npm install
npm test
\`\`\`

## 文档

- [API 文档](docs/api.md)
- [集成指南](docs/integration.md)
- [架构设计](docs/architecture.md)

## 许可证

MIT License
```

---

## .gitignore 要求

```gitignore
# Binaries
bin/
dist/
*.exe
*.dll
*.so
*.dylib

# Go
server/test_coverage.*
*.out

# Node
node_modules/
npm-debug.log*
*.log

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local

# Keys (重要!)
*.key
*.pem
secrets/
```

---

## 完成标志

### 文件检查
- [ ] 所有目录结构已创建
- [ ] .github/ 配置文件完整
- [ ] Makefile 所有命令可执行
- [ ] Docker 镜像可构建
- [ ] README.md 内容完整

### 构建检查
- [ ] `make build` 成功
- [ ] `make test` 成功
- [ ] `make docker-build` 成功

### CI/CD 检查
- [ ] GitHub Actions 工作流正常
- [ ] PR 模板已创建

---

## 下一步

完成后，可以开始 **Task 002: TSS KeyGen 实现**
