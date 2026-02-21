# AgentVault

[![Go Tests](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/go-test.yml/badge.svg)]
[![TS Tests](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/ts-test.yml/badge.svg)]
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)]

> MPC Wallet SDK for AI Agents — Give your agents the keys to autonomy

## 特性

- 🔐 **2-of-2 TSS** — 阈值签名，私钥永不完整暴露
- 🤖 **Agent-Native** — API-first 设计，无需 UI
- 🛡️ **Policy Engine** — 限额、白名单、时间窗口等策略控制
- 🚀 **5 分钟集成** — 简单的 TypeScript SDK
- 🐳 **Docker Ready** — 一键部署，可自托管

## 快速开始

### 安装 SDK

\`\`\`bash
npm install @agent-vault/sdk
\`\`\`

### 使用示例

\`\`\`typescript
import { AgentVault } from '@agent-vault/sdk';

const vault = new AgentVault({
  baseURL: 'http://localhost:8080',
  apiKey: process.env.AGENT_VAULT_API_KEY,
});

// 创建新钱包
const address = await vault.create();
console.log('Wallet address:', address);

// 签名交易
const signature = await vault.signTransaction({
  to: '0x...',
  value: '1000000'
});
\`\`\`

### 服务端部署

\`\`\`bash
docker run -d \\
  -p 8080:8080 \\
  -e VAULT_API_KEYS=your-key \\
  -e SHARD_ENCRYPTION_KEY=$(openssl rand -base64 32) \\
  ghcr.io/echowang1/agent-vault:latest
\`\`\`

## 文档

- [API 文档](docs/api.md)
- [集成指南](docs/integration.md)
- [架构设计](docs/architecture.md)

## 许可证

MIT License © 2026 [Your Name]
