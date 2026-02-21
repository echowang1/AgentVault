import type { Address } from '../types';
import type { WalletData, WalletStorage } from '../wallet';

export class MemoryWalletStorage implements WalletStorage {
  private readonly store = new Map<Address, WalletData>();

  async save(address: Address, data: WalletData): Promise<void> {
    this.store.set(address, { ...data });
  }

  async load(address: Address): Promise<WalletData | null> {
    const found = this.store.get(address);
    return found ? { ...found } : null;
  }

  async remove(address: Address): Promise<void> {
    this.store.delete(address);
  }
}
