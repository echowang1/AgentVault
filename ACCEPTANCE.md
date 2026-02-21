# Task 007: SDK 基础 + HTTP 客户端

**负责人**: Codex
**审核人**: Claude
**预计时间**: 3-4 小时
**依赖**: Task 004

---

## 功能要求

### 必须实现 (Must Have)
- [ ] TypeScript HTTP 客户端
- [ ] 类型安全的 API 封装
- [ ] 错误处理
- [ ] 支持自定义 base URL
- [ ] 支持自定义 timeout
- [ ] 完整的类型定义
- [ ] 单元测试（mock HTTP）

### 建议实现 (Should Have)
- [ ] 请求重试机制
- [ ] 请求/响应日志
- [ ] 浏览器环境支持

---

## 项目结构

```
sdk/
├── src/
│   ├── client.ts          # HTTP 客户端主类
│   ├── types.ts           # TypeScript 类型定义
│   ├── errors.ts          # 自定义错误类
│   ├── utils.ts           # 工具函数
│   └── index.ts           # 导出
├── package.json
├── tsconfig.json
├── tsconfig.build.json
└── vitest.config.ts       # 测试配置
```

---

## 类型定义

```typescript
// sdk/src/types.ts

/**
 * 钱包地址
 */
export type Address = `0x${string}`;

/**
 * 32 字节的哈希值
 */
export type Hash = `0x${string}`;

/**
 * 私钥分片（base64 编码）
 */
export type Shard = string;

/**
 * 链 ID
 */
export type ChainId = number;

/**
 * 创建钱包请求
 */
export interface CreateWalletRequest {
  chainId?: ChainId;
}

/**
 * 创建钱包响应
 */
export interface CreateWalletResponse {
  address: Address;
  publicKey: string;
  shard1: Shard;
  shard2Id: string;
}

/**
 * 签名请求
 */
export interface SignRequest {
  address: Address;
  messageHash: Hash;
  shard1: Shard;
}

/**
 * 签名响应
 */
export interface SignResponse {
  signature: Hash;
  r: string;
  s: string;
  v: number;
}

/**
 * 钱包信息
 */
export interface WalletInfo {
  address: Address;
  publicKey: string;
  createdAt: string;
}

/**
 * 策略定义
 */
export interface Policy {
  walletId: string;
  singleTxLimit?: string;
  dailyLimit?: string;
  whitelist?: Address[];
  dailyTxLimit?: number;
  startTime?: string;
  endTime?: string;
}

/**
 * 每日使用情况
 */
export interface DailyUsage {
  date: string; // YYYY-MM-DD
  totalAmount: string;
  txCount: number;
}

/**
 * API 错误响应
 */
export interface APIError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

/**
 * API 响应（成功）
 */
export interface APIResponse<T> {
  success: true;
  data: T;
}

/**
 * API 响应（失败）
 */
export interface APIErrorResponse {
  success: false;
  error: APIError;
}

/**
 * 客户端配置
 */
export interface MPCClientConfig {
  /**
   * API 基础 URL
   */
  baseURL: string;

  /**
   * API Key
   */
  apiKey: string;

  /**
   * 请求超时（毫秒）
   * @default 30000
   */
  timeout?: number;

  /**
   * 自定义 fetch 实现
   */
  fetch?: typeof fetch;
}
```

---

## HTTP 客户端

```typescript
// sdk/src/client.ts

import type {
  MPCClientConfig,
  CreateWalletRequest,
  CreateWalletResponse,
  SignRequest,
  SignResponse,
  WalletInfo,
  Policy,
  DailyUsage,
  APIResponse,
  APIErrorResponse,
} from './types';
import { MPCError, NetworkError, ValidationError } from './errors';

/**
 * MPC Wallet HTTP 客户端
 */
export class MPCClient {
  private readonly baseURL: string;
  private readonly apiKey: string;
  private readonly timeout: number;
  private readonly fetch: typeof fetch;

  constructor(config: MPCClientConfig) {
    this.baseURL = config.baseURL.replace(/\/$/, '');
    this.apiKey = config.apiKey;
    this.timeout = config.timeout ?? 30000;
    this.fetch = config.fetch ?? globalThis.fetch;
  }

  /**
   * 创建新钱包
   */
  async createWallet(req?: CreateWalletRequest): Promise<CreateWalletResponse> {
    const response = await this.post<CreateWalletResponse>(
      '/api/v1/wallet/create',
      req ?? {},
    );
    return response;
  }

  /**
   * 签名消息
   */
  async sign(req: SignRequest): Promise<SignResponse> {
    this.validateSignRequest(req);
    const response = await this.post<SignResponse>(
      '/api/v1/wallet/sign',
      req,
    );
    return response;
  }

  /**
   * 获取钱包信息
   */
  async getWallet(address: string): Promise<WalletInfo> {
    return this.get<WalletInfo>(`/api/v1/wallet/${address}`);
  }

  /**
   * 设置钱包策略
   */
  async setPolicy(address: string, policy: Partial<Policy>): Promise<void> {
    await this.put<void>(`/api/v1/wallet/${address}/policy`, policy);
  }

  /**
   * 获取钱包策略
   */
  async getPolicy(address: string): Promise<Policy> {
    return this.get<Policy>(`/api/v1/wallet/${address}/policy`);
  }

  /**
   * 获取每日使用情况
   */
  async getDailyUsage(address: string): Promise<DailyUsage> {
    return this.get<DailyUsage>(`/api/v1/wallet/${address}/usage`);
  }

  /**
   * 健康检查
   */
  async healthCheck(): Promise<{ status: string; version: string }> {
    return this.get('/health');
  }

  // ========== 私有方法 ==========

  private async get<T>(path: string): Promise<T> {
    return this.request<T>('GET', path, undefined);
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('POST', path, body);
  }

  private async put<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('PUT', path, body);
  }

  private async request<T>(
    method: string,
    path: string,
    body: unknown,
  ): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const headers = {
      'Authorization': `Bearer ${this.apiKey}`,
      'Content-Type': 'application/json',
    };

    const options: RequestInit = {
      method,
      headers,
      signal: AbortSignal.timeout(this.timeout),
    };

    if (body !== undefined) {
      options.body = JSON.stringify(body);
    }

    try {
      const response = await this.fetch(url, options);
      await this.checkResponse(response);
      const data = await response.json();
      return (data as APIResponse<T>).data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  private async checkResponse(response: Response): Promise<void> {
    if (!response.ok) {
      const data = await response.json() as APIErrorResponse;
      throw new MPCError(data.error.code, data.error.message, data.error.details);
    }
  }

  private handleError(error: unknown): Error {
    if (error instanceof MPCError) {
      return error;
    }

    if (error instanceof TypeError) {
      return new NetworkError(error.message);
    }

    if (error instanceof Error) {
      return error;
    }

    return new Error(String(error));
  }

  private validateSignRequest(req: SignRequest): void {
    // 验证地址格式
    if (!/^0x[a-fA-F0-9]{40}$/.test(req.address)) {
      throw new ValidationError('Invalid address format');
    }

    // 验证哈希格式
    if (!/^0x[a-fA-F0-9]{64}$/.test(req.messageHash)) {
      throw new ValidationError('Invalid message hash format');
    }

    // 验证 shard1
    if (!req.shard1 || req.shard1.length === 0) {
      throw new ValidationError('shard1 is required');
    }
  }
}
```

---

## 错误处理

```typescript
// sdk/src/errors.ts

/**
 * MPC 客户端基础错误类
 */
export class MPCError extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly details?: Record<string, unknown>,
  ) {
    super(message);
    this.name = 'MPCError';
  }
}

/**
 * 网络错误
 */
export class NetworkError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'NetworkError';
  }
}

/**
 * 验证错误
 */
export class ValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ValidationError';
  }
}

/**
 * 认证错误
 */
export class AuthError extends Error {
  constructor(message: string = 'Authentication failed') {
    super(message);
    this.name = 'AuthError';
  }
}

/**
 * 策略错误
 */
export class PolicyError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'PolicyError';
  }
}
```

---

## 导出

```typescript
// sdk/src/index.ts

export { MPCClient } from './client';
export type {
  MPCClientConfig,
  CreateWalletRequest,
  CreateWalletResponse,
  SignRequest,
  SignResponse,
  WalletInfo,
  Policy,
  DailyUsage,
  Address,
  Hash,
  Shard,
  ChainId,
} from './types';

export {
  MPCError,
  NetworkError,
  ValidationError,
  AuthError,
  PolicyError,
} from './errors';
```

---

## 测试用例

```typescript
// sdk/src/client.test.ts

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MPCClient } from './client';
import type { CreateWalletResponse } from './types';

describe('MPCClient', () => {
  let client: MPCClient;
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    client = new MPCClient({
      baseURL: 'http://localhost:8080',
      apiKey: 'test-api-key',
      fetch: mockFetch,
    });
  });

  describe('createWallet', () => {
    it('should create a wallet successfully', async () => {
      const mockResponse: CreateWalletResponse = {
        address: '0x1234567890123456789012345678901234567890',
        publicKey: '0x...',
        shard1: 'base64data',
        shard2Id: 'shard-2-id',
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          data: mockResponse,
        }),
      } as Response);

      const result = await client.createWallet();

      expect(result).toEqual(mockResponse);
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/wallet/create',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Authorization': 'Bearer test-api-key',
          }),
        }),
      );
    });

    it('should handle errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          success: false,
          error: {
            code: 'UNAUTHORIZED',
            message: 'Invalid API key',
          },
        }),
      } as Response);

      await expect(client.createWallet()).rejects.toThrow('Invalid API key');
    });
  });

  describe('sign', () => {
    it('should validate address format', async () => {
      await expect(
        client.sign({
          address: 'invalid' as any,
          messageHash: '0x' + 'a'.repeat(64) as any,
          shard1: 'test',
        }),
      ).rejects.toThrow('Invalid address format');
    });

    it('should validate hash format', async () => {
      await expect(
        client.sign({
          address: '0x1234567890123456789012345678901234567890' as any,
          messageHash: 'invalid' as any,
          shard1: 'test',
        }),
      ).rejects.toThrow('Invalid message hash format');
    });
  });
});
```

---

## 完成标志

### 功能验证
- [ ] 所有 API 方法可正常调用
- [ ] 类型定义完整
- [ ] 错误处理正确
- [ ] 所有测试通过

### 代码质量
- [ ] `npm test` 通过
- [ ] `npm run lint` 通过
- [ ] 测试覆盖率 > 70%

### 构建验证
- [ ] `npm run build` 成功
- [ ] 生成的 `.d.ts` 文件正确

### 发布准备
- [ ] package.json 完整
- [ ] README.md 有使用说明

---

## 下一步

完成后，可以开始 **Task 008: MPCWallet 类**
