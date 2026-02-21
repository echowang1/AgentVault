# Task 008: MPCWallet 类

**负责人**: Codex
**审核人**: Claude
**预计时间**: 4-6 小时
**依赖**: Task 007

---

## 功能要求

### 必须实现 (Must Have)
- [ ] MPCWallet 主类封装
- [ ] 本地管理 Shard 1
- [ ] ethers.js Signer 兼容接口
- [ ] 支持创建和连接现有钱包
- [ ] 签名交易和消息
- [ ] 完整的类型定义
- [ ] 单元测试

### 建议实现 (Should Have)
- [ ] viem 兼容
- [ ] 事件监听（交易签名前/后）
- [ ] 本地缓存钱包信息

---

## 接口定义

```typescript
// sdk/src/wallet.ts

import type { Signer } from 'ethers';
import type { MPCClient, Address, Hash, Shard } from './client';

/**
 * 钱包配置
 */
export interface WalletConfig {
  /**
   * MPC 客户端配置
   */
  client: {
    baseURL: string;
    apiKey: string;
    timeout?: number;
  };

  /**
   * 钱包存储（可选）
   * 用于持久化 Shard 1
   */
  storage?: WalletStorage;
}

/**
 * 钱包存储接口
 */
export interface WalletStorage {
  /**
   * 保存钱包数据
   */
  save(address: Address, data: WalletData): Promise<void>;

  /**
   * 加载钱包数据
   */
  load(address: Address): Promise<WalletData | null>;

  /**
   * 删除钱包数据
   */
  remove(address: Address): Promise<void>;
}

/**
 * 钱包持久化数据
 */
export interface WalletData {
  address: Address;
  shard1: Shard;
  publicKey: string;
  chainId?: number;
}

/**
 * 交易请求（兼容 ethers.js）
 */
export interface TransactionRequest {
  to?: Address;
  from?: Address;
  value?: string | bigint;
  data?: string;
  gasLimit?: string | bigint;
  gasPrice?: string | bigint;
  maxFeePerGas?: string | bigint;
  maxPriorityFeePerGas?: string | bigint;
  nonce?: string | number;
  chainId?: number;
}

/**
 * 签名选项
 */
export interface SignOptions {
  /**
   * 是否在签名前检查策略
   * @default true
   */
  checkPolicy?: boolean;
}

/**
 * 事件类型
 */
export type WalletEventType =
  | 'beforeSign'
  | 'afterSign'
  | 'policyCheck'
  | 'error';

/**
 * 事件监听器
 */
export type WalletEventListener = (
  event: WalletEvent,
) => void | Promise<void>;

/**
 * 钱包事件
 */
export interface WalletEvent {
  type: WalletEventType;
  data?: Record<string, unknown>;
  timestamp: Date;
}

/**
 * MPC 钱包主类
 */
export class MPCWallet implements Partial<Signer> {
  private readonly client: MPCClient;
  private readonly storage?: WalletStorage;
  private address?: Address;
  private shard1?: Shard;
  private publicKey?: string;
  private chainId?: number;
  private listeners: Map<WalletEventType, WalletEventListener[]>;

  /**
   * 创建钱包实例
   */
  constructor(config: WalletConfig);

  /**
   * 创建新钱包
   * @returns 钱包地址
   */
  create(chainId?: number): Promise<Address>;

  /**
   * 连接现有钱包
   * @param address 钱包地址
   * @param shard1 密钥碎片 1
   */
  connect(address: Address, shard1: Shard): Promise<void>;

  /**
   * 从存储恢复钱包
   * @param address 钱包地址
   */
  async load(address: Address): Promise<boolean>;

  /**
   * 断开连接
   */
  disconnect(): void;

  /**
   * 获取钱包地址
   */
  getAddress(): Address | undefined;

  /**
   * 获取链 ID
   */
  getChainId(): number | undefined;

  /**
   * 签名交易
   * @param tx 交易请求
   * @param options 签名选项
   * @returns 签名后的交易哈希
   */
  signTransaction(tx: TransactionRequest, options?: SignOptions): Promise<string>;

  /**
   * 签名消息
   * @param message 消息
   * @returns 签名
   */
  signMessage(message: string | Uint8Array): Promise<string>;

  /**
   * 签名哈希
   * @param hash 消息哈希
   * @returns 签名
   */
  signHash(hash: Hash): Promise<string>;

  /**
   * 设置策略
   * @param policy 策略配置
   */
  setPolicy(policy: Partial<Policy>): Promise<void>;

  /**
   * 获取策略
   */
  getPolicy(): Promise<Policy | null>;

  /**
   * 获取每日使用情况
   */
  getDailyUsage(): Promise<DailyUsage | null>;

  /**
   * 添加事件监听器
   */
  on(event: WalletEventType, listener: WalletEventListener): void;

  /**
   * 移除事件监听器
   */
  off(event: WalletEventType, listener: WalletEventListener): void;

  /**
   * ethers.js Signer 兼容方法
   */
  readonly provider: unknown; // 可选，用于 ethers 兼容
}

// 导出
export type {
  WalletConfig,
  WalletStorage,
  WalletData,
  TransactionRequest,
  SignOptions,
  WalletEvent,
  WalletEventListener,
};
```

---

## 实现示例

```typescript
// sdk/src/wallet.ts (实现部分)

import { MPCClient } from './client';
import type { Address, Hash, Shard, Policy, DailyUsage } from './client';
import { keccak256, toUtf8Bytes } from 'ethers';

export class MPCWallet {
  // ... 字段定义 ...

  constructor(config: WalletConfig) {
    this.client = new MPCClient(config.client);
    this.storage = config.storage;
    this.listeners = new Map();
  }

  /**
   * 创建新钱包
   */
  async create(chainId?: number): Promise<Address> {
    const response = await this.client.createWallet({ chainId });

    this.address = response.address;
    this.shard1 = response.shard1;
    this.publicKey = response.publicKey;
    this.chainId = chainId;

    // 持久化到存储
    if (this.storage) {
      await this.storage.save(this.address, {
        address: this.address,
        shard1: this.shard1,
        publicKey: this.publicKey,
        chainId: this.chainId,
      });
    }

    return this.address;
  }

  /**
   * 连接现有钱包
   */
  async connect(address: Address, shard1: Shard): Promise<void> {
    // 验证钱包是否存在
    const info = await this.client.getWallet(address);
    if (!info) {
      throw new Error('Wallet not found');
    }

    this.address = address;
    this.shard1 = shard1;
    this.publicKey = info.publicKey;

    // 持久化
    if (this.storage) {
      await this.storage.save(address, {
        address,
        shard1,
        publicKey: info.publicKey,
      });
    }
  }

  /**
   * 从存储加载
   */
  async load(address: Address): Promise<boolean> {
    if (!this.storage) {
      throw new Error('No storage configured');
    }

    const data = await this.storage.load(address);
    if (!data) {
      return false;
    }

    this.address = data.address;
    this.shard1 = data.shard1;
    this.publicKey = data.publicKey;
    this.chainId = data.chainId;

    return true;
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    this.address = undefined;
    this.shard1 = undefined;
    this.publicKey = undefined;
  }

  /**
   * 获取地址
   */
  getAddress(): Address | undefined {
    return this.address;
  }

  /**
   * 签名交易
   */
  async signTransaction(
    tx: TransactionRequest,
    options?: SignOptions,
  ): Promise<string> {
    this.ensureConnected();

    // 触发 beforeSign 事件
    await this.emit('beforeSign', { type: 'transaction', tx });

    // 计算交易哈希
    const hash = await this.getTransactionHash(tx);

    // 签名
    const signature = await this.signHash(hash);

    // 触发 afterSign 事件
    await this.emit('afterSign', { hash, signature });

    return signature;
  }

  /**
   * 签名消息
   */
  async signMessage(message: string | Uint8Array): Promise<string> {
    this.ensureConnected();

    const bytes = typeof message === 'string' ? toUtf8Bytes(message) : message;
    const hash = keccak256(bytes) as Hash;

    return this.signHash(hash);
  }

  /**
   * 签名哈希
   */
  async signHash(hash: Hash): Promise<string> {
    this.ensureConnected();
    if (!this.shard1) {
      throw new Error('No shard1 available');
    }

    const response = await this.client.sign({
      address: this.address!,
      messageHash: hash,
      shard1: this.shard1,
    });

    return response.signature;
  }

  /**
   * 设置策略
   */
  async setPolicy(policy: Partial<Policy>): Promise<void> {
    this.ensureConnected();
    await this.client.setPolicy(this.address!, policy);
  }

  /**
   * 获取策略
   */
  async getPolicy(): Promise<Policy | null> {
    this.ensureConnected();
    return this.client.getPolicy(this.address!);
  }

  /**
   * 添加事件监听
   */
  on(event: WalletEventType, listener: WalletEventListener): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event)!.push(listener);
  }

  /**
   * 移除事件监听
   */
  off(event: WalletEventType, listener: WalletEventListener): void {
    const listeners = this.listeners.get(event);
    if (listeners) {
      const index = listeners.indexOf(listener);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  }

  // ========== 私有方法 ==========

  private ensureConnected(): void {
    if (!this.address || !this.shard1) {
      throw new Error('Wallet not connected. Call create() or connect() first.');
    }
  }

  private async emit(
    event: WalletEventType,
    data?: Record<string, unknown>,
  ): Promise<void> {
    const listeners = this.listeners.get(event) ?? [];
    const eventObj: WalletEvent = {
      type: event,
      data,
      timestamp: new Date(),
    };

    for (const listener of listeners) {
      await listener(eventObj);
    }
  }

  private async getTransactionHash(tx: TransactionRequest): Promise<Hash> {
    // 实现 RLP 编码和哈希计算
    // 可以使用 ethers.js 的 utils
    // ...
    return '0x...' as Hash;
  }
}
```

---

## 内存存储实现

```typescript
// sdk/src/storage/memory.ts

import type { WalletStorage, WalletData } from './wallet';
import type { Address } from './client';

/**
 * 内存存储（用于测试）
 */
export class MemoryWalletStorage implements WalletStorage {
  private readonly store = new Map<Address, WalletData>();

  async save(address: Address, data: WalletData): Promise<void> {
    this.store.set(address, data);
  }

  async load(address: Address): Promise<WalletData | null> {
    return this.store.get(address) ?? null;
  }

  async remove(address: Address): Promise<void> {
    this.store.delete(address);
  }
}
```

---

## LocalStorage 实现

```typescript
// sdk/src/storage/local.ts

import type { WalletStorage, WalletData } from './wallet';
import type { Address } from './client';

const STORAGE_PREFIX = 'mpc-wallet:';

/**
 * LocalStorage 存储（浏览器环境）
 */
export class LocalStorageWalletStorage implements WalletStorage {
  async save(address: Address, data: WalletData): Promise<void> {
    const key = `${STORAGE_PREFIX}${address}`;
    localStorage.setItem(key, JSON.stringify(data));
  }

  async load(address: Address): Promise<WalletData | null> {
    const key = `${STORAGE_PREFIX}${address}`;
    const data = localStorage.getItem(key);
    if (!data) {
      return null;
    }
    return JSON.parse(data) as WalletData;
  }

  async remove(address: Address): Promise<void> {
    const key = `${STORAGE_PREFIX}${address}`;
    localStorage.removeItem(key);
  }
}
```

---

## 测试用例

```typescript
// sdk/src/wallet.test.ts

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MPCWallet } from './wallet';
import { MemoryWalletStorage } from './storage/memory';

describe('MPCWallet', () => {
  let wallet: MPCWallet;
  let mockClient: any;

  beforeEach(() => {
    // Mock MPCClient
    mockClient = {
      createWallet: vi.fn(),
      getWallet: vi.fn(),
      sign: vi.fn(),
      setPolicy: vi.fn(),
      getPolicy: vi.fn(),
    };

    wallet = new MPCWallet({
      client: { baseURL: 'http://localhost', apiKey: 'test' },
      storage: new MemoryWalletStorage(),
    });

    // 注入 mock client
    (wallet as any).client = mockClient;
  });

  describe('create', () => {
    it('should create a new wallet', async () => {
      const mockAddress = '0x1234567890123456789012345678901234567890' as const;
      mockClient.createWallet.mockResolvedValueOnce({
        address: mockAddress,
        publicKey: '0x...',
        shard1: 'shard1-data',
        shard2Id: 'shard2-id',
      });

      const address = await wallet.create();

      expect(address).toBe(mockAddress);
      expect(wallet.getAddress()).toBe(mockAddress);
      expect(mockClient.createWallet).toHaveBeenCalledWith({});
    });

    it('should persist to storage', async () => {
      const storage = new MemoryWalletStorage();
      const walletWithStorage = new MPCWallet({
        client: { baseURL: 'http://localhost', apiKey: 'test' },
        storage,
      });
      (walletWithStorage as any).client = mockClient;

      const mockAddress = '0x1234567890123456789012345678901234567890' as const;
      mockClient.createWallet.mockResolvedValueOnce({
        address: mockAddress,
        publicKey: '0x...',
        shard1: 'shard1-data',
        shard2Id: 'shard2-id',
      });

      await walletWithStorage.create();

      const data = await storage.load(mockAddress);
      expect(data).toEqual({
        address: mockAddress,
        shard1: 'shard1-data',
        publicKey: '0x...',
      });
    });
  });

  describe('connect', () => {
    it('should connect to existing wallet', async () => {
      const mockAddress = '0x1234567890123456789012345678901234567890' as const;
      mockClient.getWallet.mockResolvedValueOnce({
        address: mockAddress,
        publicKey: '0x...',
        createdAt: '2026-02-21T00:00:00Z',
      });

      await wallet.connect(mockAddress, 'shard1-data');

      expect(wallet.getAddress()).toBe(mockAddress);
    });

    it('should throw if wallet not found', async () => {
      mockClient.getWallet.mockResolvedValueOnce(null);

      await expect(
        wallet.connect('0x1234...7890' as any, 'shard1'),
      ).rejects.toThrow('Wallet not found');
    });
  });

  describe('signMessage', () => {
    it('should sign a message', async () => {
      mockClient.createWallet.mockResolvedValueOnce({
        address: '0x1234...7890',
        publicKey: '0x...',
        shard1: 'shard1-data',
        shard2Id: 'shard2-id',
      });

      mockClient.sign.mockResolvedValueOnce({
        signature: '0xabcdef...',
        r: '0x...',
        s: '0x...',
        v: 28,
      });

      await wallet.create();
      const signature = await wallet.signMessage('Hello, World!');

      expect(signature).toBe('0xabcdef...');
    });

    it('should throw if not connected', async () => {
      await expect(wallet.signMessage('test')).rejects.toThrow('not connected');
    });
  });

  describe('events', () => {
    it('should emit beforeSign event', async () => {
      const listener = vi.fn();
      wallet.on('beforeSign', listener);

      mockClient.createWallet.mockResolvedValueOnce({
        address: '0x1234...7890',
        publicKey: '0x...',
        shard1: 'shard1-data',
        shard2Id: 'shard2-id',
      });

      mockClient.sign.mockResolvedValueOnce({
        signature: '0x...',
        r: '0x...',
        s: '0x...',
        v: 28,
      });

      await wallet.create();
      await wallet.signMessage('test');

      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'beforeSign',
        }),
      );
    });
  });
});
```

---

## 完成标志

### 功能验证
- [ ] 创建钱包功能正常
- [ ] 连接钱包功能正常
- [ ] 签名功能正常
- [ ] 存储功能正常
- [ ] 事件系统正常
- [ ] 所有测试通过

### 代码质量
- [ ] `npm test` 通过
- [ ] `npm run lint` 通过
- [ ] 测试覆盖率 > 70%

### 构建验证
- [ ] `npm run build` 成功
- [ ] 类型导出正确

---

## 下一步

完成后，可以开始 **Task 009: 示例代码**
