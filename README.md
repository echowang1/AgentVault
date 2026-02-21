# Agent MPC Wallet

[![Go Tests](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/go-test.yml/badge.svg)]
[![TS Tests](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/ts-test.yml/badge.svg)]
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)]

> 开源的 AI Agent MPC 钱包 SDK

## 特性

- 🔐 2-of-2 TSS 阈值签名
- 🚀 5 分钟快速集成
- 📡 HTTP/REST API
- 🔌 TypeScript SDK
- 🐳 Docker 一键部署

## 快速开始

### 服务端

```bash
make build
make test
make docker-build
```

### SDK

```bash
cd sdk
npm install
npm test
```

## 文档

- [API 文档](docs/api.md)
- [集成指南](docs/integration.md)
- [架构设计](docs/architecture.md)

## 许可证

MIT License
