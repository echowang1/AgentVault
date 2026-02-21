export const PACKAGE_NAME = '@agent-vault/sdk';
export const VERSION = '0.1.0-alpha';

export interface WalletConfig {
  baseURL: string;
  apiKey: string;
}

export interface SignRequest {
  address: string;
  messageHash: string;
  shard1: string;
}

export interface SignResponse {
  signature: string;
  r: string;
  s: string;
  v: number;
}
