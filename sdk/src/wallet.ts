import {
  Transaction,
  concat,
  hexlify,
  keccak256,
  toUtf8Bytes,
} from 'ethers';

import { MPCClient } from './client';
import type {
  Address,
  DailyUsage,
  Hash,
  MPCClientConfig,
  Policy,
  Shard,
} from './types';

export interface WalletConfig {
  client: MPCClientConfig;
  storage?: WalletStorage;
}

export interface WalletStorage {
  save(address: Address, data: WalletData): Promise<void>;
  load(address: Address): Promise<WalletData | null>;
  remove(address: Address): Promise<void>;
}

export interface WalletData {
  address: Address;
  shard1: Shard;
  publicKey: string;
  chainId?: number;
}

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

export interface SignOptions {
  checkPolicy?: boolean;
}

export type WalletEventType = 'beforeSign' | 'afterSign' | 'policyCheck' | 'error';

export interface WalletEvent {
  type: WalletEventType;
  data?: Record<string, unknown>;
  timestamp: Date;
}

export type WalletEventListener = (event: WalletEvent) => void | Promise<void>;

export class MPCWallet {
  readonly provider: unknown = undefined;

  private client: MPCClient;
  private readonly storage?: WalletStorage;
  private address?: Address;
  private shard1?: Shard;
  private publicKey?: string;
  private chainId?: number;
  private readonly listeners = new Map<WalletEventType, WalletEventListener[]>();

  constructor(config: WalletConfig) {
    this.client = new MPCClient(config.client);
    this.storage = config.storage;
  }

  async create(chainId?: number): Promise<Address> {
    const response = await this.client.createWallet(chainId === undefined ? undefined : { chainId });

    this.address = response.address;
    this.shard1 = response.shard1;
    this.publicKey = response.publicKey;
    this.chainId = chainId;

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

  async connect(address: Address, shard1: Shard): Promise<void> {
    const info = await this.client.getWallet(address);
    if (!info) {
      throw new Error('Wallet not found');
    }

    this.address = address;
    this.shard1 = shard1;
    this.publicKey = info.publicKey;

    if (this.storage) {
      await this.storage.save(address, {
        address,
        shard1,
        publicKey: info.publicKey,
      });
    }
  }

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

  disconnect(): void {
    this.address = undefined;
    this.shard1 = undefined;
    this.publicKey = undefined;
    this.chainId = undefined;
  }

  getAddress(): Address | undefined {
    return this.address;
  }

  getChainId(): number | undefined {
    return this.chainId;
  }

  async signTransaction(tx: TransactionRequest, options?: SignOptions): Promise<string> {
    const hash = this.getTransactionHash(tx);
    return this.signWithContext(hash, options, { type: 'transaction', tx });
  }

  async signMessage(message: string | Uint8Array, options?: SignOptions): Promise<string> {
    const body = typeof message === 'string' ? toUtf8Bytes(message) : message;
    const prefix = toUtf8Bytes(`\x19Ethereum Signed Message:\n${body.length}`);
    const hash = keccak256(concat([prefix, body])) as Hash;
    return this.signWithContext(hash, options, {
      type: 'message',
      message: typeof message === 'string' ? message : hexlify(message),
    });
  }

  async signHash(hash: Hash, options?: SignOptions): Promise<string> {
    if (!/^0x[a-fA-F0-9]{64}$/.test(hash)) {
      throw new Error('Invalid hash format');
    }
    return this.signWithContext(hash, options, { type: 'hash' });
  }

  async setPolicy(policy: Partial<Policy>): Promise<void> {
    this.ensureConnected();
    await this.client.setPolicy(this.address!, policy);
  }

  async getPolicy(): Promise<Policy | null> {
    this.ensureConnected();
    return this.client.getPolicy(this.address!);
  }

  async getDailyUsage(): Promise<DailyUsage | null> {
    this.ensureConnected();
    return this.client.getDailyUsage(this.address!);
  }

  on(event: WalletEventType, listener: WalletEventListener): void {
    const current = this.listeners.get(event) ?? [];
    current.push(listener);
    this.listeners.set(event, current);
  }

  off(event: WalletEventType, listener: WalletEventListener): void {
    const current = this.listeners.get(event);
    if (!current) {
      return;
    }
    const index = current.indexOf(listener);
    if (index >= 0) {
      current.splice(index, 1);
    }
    this.listeners.set(event, current);
  }

  private async signWithContext(
    hash: Hash,
    options: SignOptions | undefined,
    context: Record<string, unknown>,
  ): Promise<string> {
    this.ensureConnected();

    const checkPolicy = options?.checkPolicy ?? true;
    if (checkPolicy) {
      await this.emit('policyCheck', { hash, ...context });
    }

    await this.emit('beforeSign', { hash, ...context });

    try {
      const response = await this.client.sign({
        address: this.address!,
        messageHash: hash,
        shard1: this.shard1!,
      });
      await this.emit('afterSign', { hash, signature: response.signature, ...context });
      return response.signature;
    } catch (error) {
      await this.emit('error', {
        hash,
        message: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  private ensureConnected(): void {
    if (!this.address || !this.shard1) {
      throw new Error('Wallet not connected. Call create() or connect() first.');
    }
  }

  private async emit(event: WalletEventType, data?: Record<string, unknown>): Promise<void> {
    const listeners = this.listeners.get(event) ?? [];
    const payload: WalletEvent = {
      type: event,
      data,
      timestamp: new Date(),
    };

    for (const listener of listeners) {
      await listener(payload);
    }
  }

  private getTransactionHash(tx: TransactionRequest): Hash {
    const transaction = Transaction.from({
      to: tx.to,
      value: this.toBigInt(tx.value),
      data: tx.data,
      gasLimit: this.toBigInt(tx.gasLimit),
      gasPrice: this.toBigInt(tx.gasPrice),
      maxFeePerGas: this.toBigInt(tx.maxFeePerGas),
      maxPriorityFeePerGas: this.toBigInt(tx.maxPriorityFeePerGas),
      nonce: tx.nonce === undefined ? undefined : Number(tx.nonce),
      chainId: tx.chainId ?? this.chainId,
    });

    return keccak256(transaction.unsignedSerialized) as Hash;
  }

  private toBigInt(value?: string | bigint): bigint | undefined {
    if (value === undefined) {
      return undefined;
    }
    return typeof value === 'bigint' ? value : BigInt(value);
  }
}
