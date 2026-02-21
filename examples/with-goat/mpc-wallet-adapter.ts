import { TypedDataEncoder } from 'ethers';

import { MPCWallet } from '../sdk.js';

type EVMWalletClientLike = {
  getAddress(): string;
  signTransaction(tx: Record<string, unknown>): Promise<string>;
  signMessage(message: string): Promise<string>;
  signTypedData(domain: Record<string, unknown>, types: Record<string, unknown>, value: Record<string, unknown>): Promise<string>;
};

export class GOATMPCWalletAdapter implements EVMWalletClientLike {
  constructor(private readonly wallet: MPCWallet) {}

  getAddress(): string {
    return this.wallet.getAddress() ?? '';
  }

  async signTransaction(tx: Record<string, unknown>): Promise<string> {
    return this.wallet.signTransaction({
      to: tx.to as `0x${string}` | undefined,
      value: tx.value as string | undefined,
      data: tx.data as string | undefined,
      gasLimit: tx.gas as string | undefined,
      gasPrice: tx.gasPrice as string | undefined,
      nonce: tx.nonce as number | undefined,
      chainId: tx.chainId as number | undefined,
    });
  }

  async signMessage(message: string): Promise<string> {
    return this.wallet.signMessage(message);
  }

  async signTypedData(
    domain: Record<string, unknown>,
    types: Record<string, unknown>,
    value: Record<string, unknown>,
  ): Promise<string> {
    const hash = TypedDataEncoder.hash(domain, types as Record<string, Array<{ name: string; type: string }>>, value);
    return this.wallet.signHash(hash as `0x${string}`);
  }
}

export function createGOATMPCWallet(config: {
  baseURL: string;
  apiKey: string;
  address: `0x${string}`;
  shard1: string;
}): Promise<GOATMPCWalletAdapter> {
  const wallet = new MPCWallet({
    client: {
      baseURL: config.baseURL,
      apiKey: config.apiKey,
    },
  });

  return wallet.connect(config.address, config.shard1).then(() => new GOATMPCWalletAdapter(wallet));
}
