import {
  AuthError,
  MPCError,
  NetworkError,
  PolicyError,
  ValidationError,
} from './errors';
import type {
  APIErrorResponse,
  APIResponse,
  CreateWalletRequest,
  CreateWalletResponse,
  DailyUsage,
  MPCClientConfig,
  Policy,
  SignRequest,
  SignResponse,
  WalletInfo,
} from './types';

export class MPCClient {
  private readonly baseURL: string;
  private readonly apiKey: string;
  private readonly timeout: number;
  private readonly fetchImpl?: typeof globalThis.fetch;

  constructor(config: MPCClientConfig) {
    this.baseURL = config.baseURL.replace(/\/$/, '');
    this.apiKey = config.apiKey;
    this.timeout = config.timeout ?? 30_000;
    this.fetchImpl = config.fetch ?? globalThis.fetch;
  }

  async createWallet(req?: CreateWalletRequest): Promise<CreateWalletResponse> {
    const body = req?.chainId ? { chain_id: req.chainId } : {};
    const data = await this.post<{
      address: string;
      public_key: string;
      shard1: string;
      shard2_id: string;
    }>('/api/v1/wallet/create', body);

    return {
      address: data.address as CreateWalletResponse['address'],
      publicKey: data.public_key,
      shard1: data.shard1,
      shard2Id: data.shard2_id,
    };
  }

  async sign(req: SignRequest): Promise<SignResponse> {
    this.validateAddress(req.address);
    this.validateHash(req.messageHash);
    if (!req.shard1) {
      throw new ValidationError('shard1 is required');
    }

    const data = await this.post<{
      signature: string;
      r: string;
      s: string;
      v: number;
    }>('/api/v1/wallet/sign', {
      address: req.address,
      message_hash: req.messageHash,
      shard1: req.shard1,
    });

    return {
      signature: data.signature as SignResponse['signature'],
      r: data.r,
      s: data.s,
      v: data.v,
    };
  }

  async getWallet(address: string): Promise<WalletInfo> {
    this.validateAddress(address);

    const data = await this.get<{
      address: string;
      public_key: string;
      created_at: string;
    }>(`/api/v1/wallet/${address}`);

    return {
      address: data.address as WalletInfo['address'],
      publicKey: data.public_key,
      createdAt: data.created_at,
    };
  }

  async setPolicy(address: string, policy: Partial<Policy>): Promise<void> {
    this.validateAddress(address);

    const body = {
      single_tx_limit: policy.singleTxLimit,
      daily_limit: policy.dailyLimit,
      whitelist: policy.whitelist,
      daily_tx_limit: policy.dailyTxLimit,
      start_time: policy.startTime,
      end_time: policy.endTime,
    };

    await this.put<void>(`/api/v1/wallet/${address}/policy`, body);
  }

  async getPolicy(address: string): Promise<Policy> {
    this.validateAddress(address);

    const data = await this.get<{
      wallet_id: string;
      single_tx_limit?: string;
      daily_limit?: string;
      whitelist?: string[];
      daily_tx_limit?: number;
      start_time?: string;
      end_time?: string;
    }>(`/api/v1/wallet/${address}/policy`);

    return {
      walletId: data.wallet_id,
      singleTxLimit: data.single_tx_limit,
      dailyLimit: data.daily_limit,
      whitelist: data.whitelist as Policy['whitelist'],
      dailyTxLimit: data.daily_tx_limit,
      startTime: data.start_time,
      endTime: data.end_time,
    };
  }

  async getDailyUsage(address: string): Promise<DailyUsage> {
    this.validateAddress(address);

    const data = await this.get<{
      date: string;
      total_amount: string;
      tx_count: number;
    }>(`/api/v1/wallet/${address}/usage`);

    return {
      date: data.date,
      totalAmount: data.total_amount,
      txCount: data.tx_count,
    };
  }

  async healthCheck(): Promise<{ status: string; version: string }> {
    return this.get<{ status: string; version: string }>('/health');
  }

  private async get<T>(path: string): Promise<T> {
    return this.request<T>('GET', path);
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('POST', path, body);
  }

  private async put<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('PUT', path, body);
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const fetcher = this.fetchImpl;
    if (!fetcher) {
      throw new NetworkError('fetch is not available in this runtime');
    }

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.apiKey}`,
      'Content-Type': 'application/json',
    };

    const init: RequestInit = {
      method,
      headers,
      signal: controller.signal,
      body: body === undefined ? undefined : JSON.stringify(body),
    };

    try {
      const response = await fetcher(`${this.baseURL}${path}`, init);
      if (!response.ok) {
        await this.throwFromErrorResponse(response);
      }
      const payload = (await response.json()) as APIResponse<T>;
      return payload.data;
    } catch (error) {
      throw this.handleError(error);
    } finally {
      clearTimeout(timeoutId);
    }
  }

  private async throwFromErrorResponse(response: Response): Promise<never> {
    let errorPayload: APIErrorResponse | undefined;
    try {
      errorPayload = (await response.json()) as APIErrorResponse;
    } catch {
      // ignore json parsing failure and fallback to status text
    }

    const code = errorPayload?.error?.code ?? String(response.status);
    const message = errorPayload?.error?.message ?? (response.statusText || 'request failed');
    const details = errorPayload?.error?.details;

    if (response.status === 401) {
      throw new AuthError(message);
    }
    if (code.includes('POLICY')) {
      throw new PolicyError(message);
    }
    throw new MPCError(code, message, details);
  }

  private handleError(error: unknown): Error {
    if (error instanceof MPCError || error instanceof ValidationError || error instanceof AuthError || error instanceof PolicyError) {
      return error;
    }

    if (error instanceof DOMException && error.name === 'AbortError') {
      return new NetworkError(`request timed out after ${this.timeout}ms`);
    }

    if (error instanceof TypeError) {
      return new NetworkError(error.message);
    }

    if (error instanceof Error) {
      return error;
    }

    return new Error(String(error));
  }

  private validateAddress(address: string): void {
    if (!/^0x[a-fA-F0-9]{40}$/.test(address)) {
      throw new ValidationError('Invalid address format');
    }
  }

  private validateHash(hash: string): void {
    if (!/^0x[a-fA-F0-9]{64}$/.test(hash)) {
      throw new ValidationError('Invalid message hash format');
    }
  }
}
