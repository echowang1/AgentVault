# Task 010: 端到端测试

**负责人**: Codex
**审核人**: Claude
**预计时间**: 3-4 小时
**依赖**: Task 008, 009

---

## 功能要求

### 必须实现 (Must Have)
- [ ] 完整的 E2E 测试流程
- [ ] 测试网验证（Sepolia/Base Sepolia）
- [ ] 测试真实交易发送和确认
- [ ] 测试报告生成
- [ ] CI/CD 集成

### 建议实现 (Should Have)
- [ ] 性能测试
- [ ] 压力测试
- [ ] 多链测试

---

## 测试场景

### 场景 1: 创建钱包 → 签名消息 → 验证

```
1. 启动 MPC 服务
2. 使用 SDK 创建钱包
3. 签名测试消息
4. 使用 ethers.js 验证签名
5. 检查恢复地址匹配
```

### 场景 2: 创建钱包 → 签名交易 → 发送到测试网

```
1. 启动 MPC 服务
2. 使用 SDK 创建钱包
3. 向钱包转入测试 ETH（水龙头）
4. 签名测试交易
5. 发送到测试网
6. 等待交易确认
7. 验证交易状态
```

### 场景 3: 策略引擎测试

```
1. 创建钱包
2. 设置策略（单笔限额 0.01 ETH）
3. 尝试签名 0.001 ETH 交易 → 成功
4. 尝试签名 0.1 ETH 交易 → 失败
5. 设置白名单
6. 尝试签名给白名单地址 → 成功
7. 尝试签名给非白名单地址 → 失败
```

### 场景 4: 多钱包并发测试

```
1. 创建 5 个钱包
2. 并发签名请求
3. 验证所有签名正确
4. 验证没有密钥泄露
```

---

## 测试代码

```typescript
// tests/e2e.test.ts

import { describe, it, expect, beforeAll } from 'vitest';
import { MPCWallet } from '@agent-mpc-wallet/sdk';
import { ethers, verifyMessage, recoverAddress } from 'ethers';
import { spawn } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(require('child_process').exec);

describe('MPC Wallet E2E Tests', () => {
  let serverProcess: any;
  let serverUrl: string;
  let apiKey: string = 'test-api-key-e2e';

  beforeAll(async () => {
    // 1. 启动测试服务器
    serverUrl = 'http://localhost:8080';
    serverProcess = spawn('go', ['run', '../server/cmd/server/main.go'], {
      env: {
        ...process.env,
        MPC_API_KEYS: apiKey,
        DB_PATH: ':memory:',
      },
      stdio: 'pipe',
    });

    // 等待服务器启动
    await new Promise(resolve => setTimeout(resolve, 3000));

    // 健康检查
    const response = await fetch(`${serverUrl}/health`);
    expect(response.ok).toBe(true);
  }, 30000);

  afterAll(async () => {
    if (serverProcess) {
      serverProcess.kill();
    }
  });

  describe('Scenario 1: Create Wallet and Sign Message', () => {
    it('should create wallet and sign message successfully', async () => {
      // 创建钱包
      const wallet = new MPCWallet({
        client: {
          baseURL: serverUrl,
          apiKey: apiKey,
        },
      });

      const address = await wallet.create();
      expect(address).toMatch(/^0x[a-fA-F0-9]{40}$/);

      // 签名消息
      const message = 'Hello, MPC Wallet!';
      const signature = await wallet.signMessage(message);

      expect(signature).toMatch(/^0x[a-fA-F0-9]{130}$/);

      // 验证签名
      const recovered = verifyMessage(message, signature);
      expect(recovered.toLowerCase()).toBe(address.toLowerCase());
    });

    it('should sign hash correctly', async () => {
      const wallet = new MPCWallet({
        client: {
          baseURL: serverUrl,
          apiKey: apiKey,
        },
      });

      const address = await wallet.create();

      const message = 'Test message for hash signing';
      const hash = ethers.keccak256(ethers.toUtf8Bytes(message));

      const signature = await wallet.signHash(hash as any);

      // 恢复签名者
      const recovered = recoverAddress(hash, signature);
      expect(recovered.toLowerCase()).toBe(address.toLowerCase());
    });
  });

  describe('Scenario 2: Sign and Send Transaction', () => {
    it('should sign transaction and send to testnet', async () => {
      const wallet = new MPCWallet({
        client: {
          baseURL: serverUrl,
          apiKey: apiKey,
        },
      });

      const address = await wallet.create();

      // 获取测试 ETH（如果余额不足）
      const provider = ethers.getDefaultProvider('sepolia');
      const balance = await provider.getBalance(address);

      if (balance < ethers.parseEther('0.01')) {
        console.log('⚠️  Insufficient test ETH. Please fund:', address);
        // 跳过交易发送测试
        return;
      }

      // 构建交易
      const tx = {
        to: '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb' as const,
        value: ethers.parseEther('0.001'),
        gasLimit: 21000,
        chainId: 11155111, // Sepolia
      };

      // 签名交易
      const signature = await wallet.signTransaction(tx);

      // 发送交易
      const txResponse = await provider.broadcastTransaction(signature);
      console.log('Transaction hash:', txResponse.hash);

      // 等待确认
      const receipt = await txResponse.wait();
      expect(receipt).toBeDefined();
      expect(receipt?.status).toBe(1);
    }, 60000);
  });

  describe('Scenario 3: Policy Engine', () => {
    it('should enforce single transaction limit', async () => {
      const wallet = new MPCWallet({
        client: {
          baseURL: serverUrl,
          apiKey: apiKey,
        },
      });

      const address = await wallet.create();

      // 设置策略：单笔限额 0.01 ETH
      await wallet.setPolicy({
        singleTxLimit: ethers.parseEther('0.01').toString(),
      });

      // 小额交易应该成功
      const smallTx = {
        to: '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb' as const,
        value: ethers.parseEther('0.001'),
      };

      await expect(wallet.signTransaction(smallTx)).resolves.toBeDefined();

      // 大额交易应该失败
      const largeTx = {
        to: '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb' as const,
        value: ethers.parseEther('0.1'),
      };

      await expect(wallet.signTransaction(largeTx)).rejects.toThrow();
    });

    it('should enforce whitelist', async () => {
      const wallet = new MPCWallet({
        client: {
          baseURL: serverUrl,
          apiKey: apiKey,
        },
      });

      const address = await wallet.create();

      const uniswap = '0xE592427A0AEce92De3Edee1F18E0157C05861564' as const;
      const random = '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb' as const;

      // 设置白名单
      await wallet.setPolicy({
        whitelist: [uniswap],
      });

      // 白名单地址应该成功
      const whitelistTx = {
        to: uniswap,
        value: ethers.parseEther('0.001'),
      };

      await expect(wallet.signTransaction(whitelistTx)).resolves.toBeDefined();

      // 非白名单地址应该失败
      const nonWhitelistTx = {
        to: random,
        value: ethers.parseEther('0.001'),
      };

      await expect(wallet.signTransaction(nonWhitelistTx)).rejects.toThrow();
    });
  });

  describe('Scenario 4: Concurrent Operations', () => {
    it('should handle multiple concurrent wallets', async () => {
      const walletCount = 5;
      const wallets = await Promise.all(
        Array.from({ length: walletCount }, () =>
          new MPCWallet({
            client: { baseURL: serverUrl, apiKey },
          }).create(),
        ),
      );

      expect(wallets).toHaveLength(walletCount);
      expect(new Set(wallets).size).toBe(walletCount); // 所有地址不同
    });

    it('should handle concurrent signing requests', async () => {
      const wallet = new MPCWallet({
        client: {
          baseURL: serverUrl,
          apiKey: apiKey,
        },
      });

      const address = await wallet.create();

      // 并发签名 10 个消息
      const signatures = await Promise.all(
        Array.from({ length: 10 }, (_, i) =>
          wallet.signMessage(`Message ${i}`),
        ),
      );

      expect(signatures).toHaveLength(10);

      // 验证所有签名
      signatures.forEach((sig, i) => {
        expect(sig).toMatch(/^0x[a-fA-F0-9]{130}$/);
        const recovered = verifyMessage(`Message ${i}`, sig);
        expect(recovered.toLowerCase()).toBe(address.toLowerCase());
      });
    });
  });

  describe('Scenario 5: Storage Persistence', () => {
    it('should persist and load wallet data', async () => {
      const storage = new MemoryWalletStorage();
      const wallet = new MPCWallet({
        client: { baseURL: serverUrl, apiKey },
        storage,
      });

      // 创建钱包
      const address = await wallet.create();

      // 断开
      wallet.disconnect();

      // 创建新实例并加载
      const wallet2 = new MPCWallet({
        client: { baseURL: serverUrl, apiKey },
        storage,
      });

      const loaded = await wallet2.load(address);
      expect(loaded).toBe(true);
      expect(wallet2.getAddress()).toBe(address);

      // 验证可以签名
      const signature = await wallet2.signMessage('test');
      expect(signature).toBeDefined();
    });
  });
});
```

---

## 性能测试

```typescript
// tests/performance.test.ts

import { describe, it, expect } from 'vitest';
import { MPCWallet } from '@agent-mpc-wallet/sdk';

describe('Performance Tests', () => {
  it('should create wallet in under 10 seconds', async () => {
    const start = Date.now();

    const wallet = new MPCWallet({
      client: { baseURL: 'http://localhost:8080', apiKey: 'test' },
    });

    await wallet.create();

    const elapsed = Date.now() - start;
    expect(elapsed).toBeLessThan(10000);
    console.log(`Wallet creation took ${elapsed}ms`);
  });

  it('should sign message in under 3 seconds', async () => {
    const wallet = new MPCWallet({
      client: { baseURL: 'http://localhost:8080', apiKey: 'test' },
    });

    await wallet.create();

    const start = Date.now();
    await wallet.signMessage('performance test');
    const elapsed = Date.now() - start;

    expect(elapsed).toBeLessThan(3000);
    console.log(`Signing took ${elapsed}ms`);
  });

  it('should handle 100 transactions in under 60 seconds', async () => {
    const wallet = new MPCWallet({
      client: { baseURL: 'http://localhost:8080', apiKey: 'test' },
    });

    await wallet.create();

    const start = Date.now();

    const signatures = await Promise.all(
      Array.from({ length: 100 }, (_, i) =>
        wallet.signMessage(`Message ${i}`),
      ),
    );

    const elapsed = Date.now() - start;
    expect(signatures).toHaveLength(100);
    expect(elapsed).toBeLessThan(60000);

    console.log(`100 signatures took ${elapsed}ms`);
    console.log(`Average: ${elapsed / 100}ms per signature`);
  });
});
```

---

## 测试配置

```typescript
// vitest.config.ts

import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'tests/',
        '**/*.test.ts',
        '**/*.spec.ts',
        'examples/',
      ],
    },
    setupFiles: ['./tests/setup.ts'],
  },
});
```

```typescript
// tests/setup.ts

import { beforeAll } from 'vitest';

beforeAll(() => {
  // 全局测试设置
  console.log('Starting E2E tests...');
  console.log('Make sure the MPC server is running!');
});
```

---

## CI/CD 集成

```yaml
# .github/workflows/e2e.yml

name: E2E Tests

on:
  push:
    branches: [main, task/**]
  pull_request:
    branches: [main]

jobs:
  e2e:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: test_db
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: sdk/package-lock.json

      - name: Install dependencies
        run: |
          cd server && go mod download
          cd sdk && npm ci

      - name: Build server
        run: |
          cd server && go build -o ../bin/server ./cmd/server

      - name: Start server
        run: |
          mkdir -p data
          export MPC_API_KEYS=test-key
          export DB_PATH=./data/test.db
          ./bin/server &
          sleep 5

      - name: Run E2E tests
        run: |
          cd sdk && npm run test:e2e

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./sdk/coverage/e2e coverage-final.json
```

---

## 测试脚本

```json
// package.json

{
  "scripts": {
    "test": "vitest",
    "test:unit": "vitest run --coverage",
    "test:e2e": "vitest run tests/e2e.test.ts",
    "test:perf": "vitest run tests/performance.test.ts",
    "test:all": "npm-run-all test:*"
  }
}
```

---

## 完成标志

### 测试完整性
- [ ] 所有 E2E 场景已实现
- [ ] 性能测试已实现
- [ ] 测试可以正常运行
- [ ] CI/CD 已配置

### 测试质量
- [ ] 测试覆盖率 > 70%
- [ ] 测试运行时间 < 5 分钟
- [ ] 测试有清晰的输出

### 集成验证
- [ ] 测试网交易成功
- [ ] 签名验证通过
- [ ] 策略检查正确

---

## 下一步

完成后，可以开始 **Task 011: API 文档 + 部署指南**
