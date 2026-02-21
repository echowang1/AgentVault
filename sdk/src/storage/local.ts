import type { Address } from '../types';
import type { WalletData, WalletStorage } from '../wallet';

const DEFAULT_PREFIX = 'mpc-wallet:';

export class LocalStorageWalletStorage implements WalletStorage {
  private readonly storage: Storage | undefined;
  private readonly prefix: string;

  constructor(storage?: Storage, prefix: string = DEFAULT_PREFIX) {
    this.storage = storage ?? globalThis.localStorage;
    this.prefix = prefix;
  }

  async save(address: Address, data: WalletData): Promise<void> {
    const driver = this.requireStorage();
    driver.setItem(this.key(address), JSON.stringify(data));
  }

  async load(address: Address): Promise<WalletData | null> {
    const driver = this.requireStorage();
    const raw = driver.getItem(this.key(address));
    return raw ? (JSON.parse(raw) as WalletData) : null;
  }

  async remove(address: Address): Promise<void> {
    const driver = this.requireStorage();
    driver.removeItem(this.key(address));
  }

  private key(address: Address): string {
    return `${this.prefix}${address.toLowerCase()}`;
  }

  private requireStorage(): Storage {
    if (!this.storage) {
      throw new Error('localStorage is not available in this runtime');
    }
    return this.storage;
  }
}
