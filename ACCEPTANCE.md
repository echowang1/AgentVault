# Task 009: 示例代码

**负责人**: Codex
**审核人**: Claude
**预计时间**: 2-3 小时
**依赖**: Task 008

---

## 功能要求

### 必须实现 (Must Have)
- [ ] 创建钱包示例
- [ ] 签名交易示例
- [ ] 签名消息示例
- [ ] 策略设置示例
- [ ] README 快速开始指南
- [ ] 所有示例可运行

### 建议实现 (Should Have)
- [ ] ElizaOS 插件示例
- [ ] GOAT SDK 集成示例
- [ ] 错误处理示例
- [ ] TypeScript/JavaScript 示例

---

## 示例结构

```
examples/
├── basic/
│   ├── 01-create-wallet.ts     # 创建钱包
│   ├── 02-sign-message.ts      # 签名消息
│   ├── 03-sign-transaction.ts  # 签名交易
│   ├── 04-set-policy.ts        # 设置策略
│   └── README.md
│
├── with-eliza/
│   ├── mpc-wallet-plugin.ts    # ElizaOS 插件
│   └── README.md
│
├── with-goat/
│   ├── mpc-wallet-adapter.ts   # GOAT SDK 适配器
│   └── README.md
│
└── README.md                   # 示例总览
```

---

## 基础示例

### 01-create-wallet.ts

```typescript
// examples/basic/01-create-wallet.ts

import { MPCWallet, MemoryWalletStorage } from '@agent-mpc-wallet/sdk';

async function main() {
  // 1. 创建钱包实例
  const wallet = new MPCWallet({
    client: {
      baseURL: process.env.MPC_SERVER_URL ?? 'http://localhost:8080',
      apiKey: process.env.MPC_API_KEY ?? 'test-api-key',
    },
    storage: new MemoryWalletStorage(), // 生产环境使用持久化存储
  });

  try {
    // 2. 创建新钱包
    console.log('Creating new wallet...');
    const address = await wallet.create(1); // Chain ID 1 = Ethereum

    console.log(`✅ Wallet created successfully!`);
    console.log(`   Address: ${address}`);
    console.log(`   Shard 1 (save this!): ${(wallet as any).shard1}`);

    // 3. 后续使用时，从存储加载
    // await wallet.load(address);

  } catch (error) {
    console.error('❌ Failed to create wallet:', error);
    process.exit(1);
  }
}

main();
```

### 02-sign-message.ts

```typescript
// examples/basic/02-sign-message.ts

import { MPCWallet } from '@agent-mpc-wallet/sdk';

async function main() {
  const wallet = new MPCWallet({
    client: {
      baseURL: process.env.MPC_SERVER_URL ?? 'http://localhost:8080',
      apiKey: process.env.MPC_API_KEY ?? 'test-api-key',
    },
  });

  // 连接已有钱包
  const address = '0x...' as const;
  const shard1 = 'base64...';

  await wallet.connect(address, shard1);

  try {
    // 签名消息
    const message = 'Hello, Agent MPC Wallet!';
    console.log(`Signing message: "${message}"`);

    const signature = await wallet.signMessage(message);

    console.log(`✅ Message signed successfully!`);
    console.log(`   Signature: ${signature}`);

    // 验证签名（使用 ethers.js）
    // const recovered = verifyMessage(message, signature);
    // console.log(`   Recovered address: ${recovered}`);

  } catch (error) {
    console.error('❌ Failed to sign message:', error);
  }
}

main();
```

### 03-sign-transaction.ts

```typescript
// examples/basic/03-sign-transaction.ts

import { MPCWallet } from '@agent-mpc-wallet/sdk';
import { parseEther } from 'ethers';

async function main() {
  const wallet = new MPCWallet({
    client: {
      baseURL: process.env.MPC_SERVER_URL ?? 'http://localhost:8080',
      apiKey: process.env.MPC_API_KEY ?? 'test-api-key',
    },
  });

  // 连接钱包
  const address = '0x...' as const;
  const shard1 = 'base64...';
  await wallet.connect(address, shard1);

  try {
    // 构建交易
    const tx = {
      to: '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb' as const, // 示例地址
      value: parseEther('0.01'), // 0.001 ETH
      gasLimit: 21000,
      chainId: 1,
    };

    console.log('Signing transaction:');
    console.log(`   To: ${tx.to}`);
    console.log(`   Value: ${tx.value} wei`);

    // 签名
    const signature = await wallet.signTransaction(tx);

    console.log(`✅ Transaction signed successfully!`);
    console.log(`   Signature: ${signature}`);

    // 广播交易（使用 ethers.js provider）
    // const txHash = await provider.sendTransaction(signature);
    // console.log(`   Tx Hash: ${txHash}`);

  } catch (error) {
    console.error('❌ Failed to sign transaction:', error);

    // 检查是否是策略错误
    if (error instanceof PolicyError) {
      console.error('   Transaction rejected by policy');
    }
  }
}

main();
```

### 04-set-policy.ts

```typescript
// examples/basic/04-set-policy.ts

import { MPCWallet } from '@agent-mpc-wallet/sdk';
import { parseEther } from 'ethers';

async function main() {
  const wallet = new MPCWallet({
    client: {
      baseURL: process.env.MPC_SERVER_URL ?? 'http://localhost:8080',
      apiKey: process.env.MPC_API_KEY ?? 'test-api-key',
    },
  });

  const address = '0x...' as const;
  const shard1 = 'base64...';
  await wallet.connect(address, shard1);

  try {
    // 设置策略
    console.log('Setting wallet policy...');

    await wallet.setPolicy({
      singleTxLimit: parseEther('1').toString(), // 单笔最多 1 ETH
      dailyLimit: parseEther('10').toString(),   // 每日最多 10 ETH
      whitelist: [
        '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb', // Uniswap Router
        '0xE592427A0AEce92De3Edee1F18E0157C05861564', // Uniswap V3 Router
      ],
      dailyTxLimit: 100, // 每日最多 100 笔交易
    });

    console.log('✅ Policy set successfully!');

    // 查询策略
    const policy = await wallet.getPolicy();
    console.log('Current policy:');
    console.log(`   Single tx limit: ${policy?.singleTxLimit} wei`);
    console.log(`   Daily limit: ${policy?.dailyLimit} wei`);
    console.log(`   Whitelist: ${policy?.whitelist?.join(', ')}`);

    // 查询每日使用
    const usage = await wallet.getDailyUsage();
    console.log('Today\'s usage:');
    console.log(`   Total amount: ${usage?.totalAmount} wei`);
    console.log(`   Tx count: ${usage?.txCount}`);

  } catch (error) {
    console.error('❌ Failed to set policy:', error);
  }
}

main();
```

---

## ElizaOS 插件示例

```typescript
// examples/with-eliza/mpc-wallet-plugin.ts

import { Plugin, IAgentRuntime } from '@elizaos/core';
import { MPCWallet } from '@agent-mpc-wallet/sdk';

/**
 * MPC Wallet Plugin for ElizaOS
 *
 * 使用方式:
 * 1. 在 ElizaOS 配置中添加插件
 * 2. 设置环境变量 MPC_SERVER_URL 和 MPC_API_KEY
 * 3. Agent 可以使用钱包进行签名
 */
export class MPCWalletPlugin implements Plugin {
  name = 'mpc-wallet';
  description = 'MPC wallet for secure signing';

  private wallet?: MPCWallet;

  async init(runtime: IAgentRuntime): Promise<void> {
    // 从配置读取 MPC 服务地址
    const serverUrl = runtime.getSetting('MPC_SERVER_URL');
    const apiKey = runtime.getSetting('MPC_API_KEY');

    if (!serverUrl || !apiKey) {
      throw new Error('MPC_SERVER_URL and MPC_API_KEY must be set');
    }

    // 创建钱包实例
    this.wallet = new MPCWallet({
      client: {
        baseURL: serverUrl,
        apiKey: apiKey,
      },
    });

    // 连接或创建钱包
    const address = runtime.getSetting('WALLET_ADDRESS');
    const shard1 = runtime.getSetting('WALLET_SHARD1');

    if (address && shard1) {
      await this.wallet.connect(address as any, shard1);
      console.log(`[MPC Wallet] Connected to ${address}`);
    } else {
      const newAddress = await this.wallet.create();
      console.log(`[MPC Wallet] Created new wallet: ${newAddress}`);
      console.log(`[MPC Wallet] IMPORTANT: Save your shard1! ${(this.wallet as any).shard1}`);
    }

    // 注册 Actions
    runtime.registerAction({
      name: 'SIGN_TRANSACTION',
      description: 'Sign a transaction using MPC wallet',
      handler: async (args: any) => {
        if (!this.wallet) {
          throw new Error('Wallet not initialized');
        }

        const signature = await this.wallet.signTransaction({
          to: args.to,
          value: args.value,
          data: args.data,
          gasLimit: args.gasLimit,
        });

        return {
          success: true,
          signature,
        };
      },
    });

    runtime.registerAction({
      name: 'SIGN_MESSAGE',
      description: 'Sign a message using MPC wallet',
      handler: async (args: any) => {
        if (!this.wallet) {
          throw new Error('Wallet not initialized');
        }

        const signature = await this.wallet.signMessage(args.message);

        return {
          success: true,
          signature,
        };
      },
    });

    runtime.registerAction({
      name: 'GET_WALLET_ADDRESS',
      description: 'Get the wallet address',
      handler: async () => {
        if (!this.wallet) {
          throw new Error('Wallet not initialized');
        }

        return {
          success: true,
          address: this.wallet.getAddress(),
        };
      },
    });
  }

  async cleanup(): Promise<void> {
    this.wallet?.disconnect();
  }
}

export default MPCWalletPlugin;
```

---

## GOAT SDK 适配器示例

```typescript
// examples/with-goat/mpc-wallet-adapter.ts

import { EVMWalletClient } from '@goat-sdk/wallet';
import { MPCWallet } from '@agent-mpc-wallet/sdk';

/**
 * MPC Wallet Adapter for GOAT SDK
 *
 * 实现 GOAT 的 EVMWalletClient 接口
 * 使得 GOAT 框架可以使用 MPC 钱包
 */
export class GOATMPCWalletAdapter implements EVMWalletClient {
  private wallet: MPCWallet;

  constructor(wallet: MPCWallet) {
    this.wallet = wallet;
  }

  /**
   * 获取钱包地址
   */
  getAddress(): string {
    return this.wallet.getAddress() ?? '';
  }

  /**
   * 签名交易
   */
  async signTransaction(tx: any): Promise<string> {
    return this.wallet.signTransaction({
      to: tx.to,
      value: tx.value,
      data: tx.data,
      gasLimit: tx.gas,
      gasPrice: tx.gasPrice,
      nonce: tx.nonce,
      chainId: tx.chainId,
    });
  }

  /**
   * 签名消息
   */
  async signMessage(message: string): Promise<string> {
    return this.wallet.signMessage(message);
  }

  /**
   * 签名 typed data (EIP-712)
   */
  async signTypedData(domain: any, types: any, value: any): Promise<string> {
    // 实现 EIP-712 签名
    // 需要先计算 hash，然后调用 signHash
    const hash = ''; // TODO: 计算 EIP-712 hash
    return this.wallet.signHash(hash as any);
  }
}

// 使用示例
export function createGOATMPCWallet(config: any): EVMWalletClient {
  const wallet = new MPCWallet({
    client: {
      baseURL: config.serverURL,
      apiKey: config.apiKey,
    },
  });

  return new GOATMPCWalletAdapter(wallet);
}
```

---

## README 文档

### examples/README.md

```markdown
# Agent MPC Wallet - 示例代码

本目录包含 MPC Wallet SDK 的使用示例。

## 环境设置

```bash
# 安装依赖
npm install

# 设置环境变量
export MPC_SERVER_URL=http://localhost:8080
export MPC_API_KEY=your-api-key
```

## 基础示例

### 1. 创建钱包

```bash
npm run example:create
```

### 2. 签名消息

```bash
npm run example:sign-message
```

### 3. 签名交易

```bash
npm run example:sign-tx
```

### 4. 设置策略

```bash
npm run example:set-policy
```

## 框架集成

### ElizaOS 插件

参见 [with-eliza/README.md](./with-eliza/README.md)

### GOAT SDK 适配器

参见 [with-goat/README.md](./with-goat/README.md)

## 运行所有示例

```bash
npm run examples
```
```

---

## package.json 脚本

```json
{
  "scripts": {
    "example:create": "tsx examples/basic/01-create-wallet.ts",
    "example:sign-message": "tsx examples/basic/02-sign-message.ts",
    "example:sign-tx": "tsx examples/basic/03-sign-transaction.ts",
    "example:set-policy": "tsx examples/basic/04-set-policy.ts",
    "examples": "npm-run-all example:*"
  }
}
```

---

## 完成标志

### 文件完整性
- [ ] 所有基础示例已创建
- [ ] ElizaOS 插件示例已创建
- [ ] GOAT SDK 适配器示例已创建
- [ ] README 文档完整

### 可运行性
- [ ] 所有示例可成功编译
- [ ] 所有示例可正常运行（需要 mock 或真实服务）
- [ ] 有清晰的输出说明

### 文档质量
- [ ] 代码有充分注释
- [ ] README 有使用说明
- [ ] 有环境变量说明

---

## 下一步

完成后，可以开始 **Task 010: 端到端测试**
