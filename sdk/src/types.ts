export type Address = `0x${string}`;
export type Hash = `0x${string}`;
export type Shard = string;
export type ChainId = number;

export interface CreateWalletRequest {
  chainId?: ChainId;
}

export interface CreateWalletResponse {
  address: Address;
  publicKey: string;
  shard1: Shard;
  shard2Id: string;
}

export interface SignRequest {
  address: Address;
  messageHash: Hash;
  shard1: Shard;
}

export interface SignResponse {
  signature: Hash;
  r: string;
  s: string;
  v: number;
}

export interface WalletInfo {
  address: Address;
  publicKey: string;
  createdAt: string;
}

export interface Policy {
  walletId: string;
  singleTxLimit?: string;
  dailyLimit?: string;
  whitelist?: Address[];
  dailyTxLimit?: number;
  startTime?: string;
  endTime?: string;
}

export interface DailyUsage {
  date: string;
  totalAmount: string;
  txCount: number;
}

export interface APIError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface APIResponse<T> {
  success: true;
  data: T;
}

export interface APIErrorResponse {
  success: false;
  error: APIError;
}

export interface MPCClientConfig {
  baseURL: string;
  apiKey: string;
  timeout?: number;
  fetch?: typeof globalThis.fetch;
}
