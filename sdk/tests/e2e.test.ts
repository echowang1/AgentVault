import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import {
  keccak256,
  recoverAddress,
  toUtf8Bytes,
  verifyMessage,
  type SignatureLike,
} from 'ethers';

import {
  MPCClient,
  MPCWallet,
  MemoryWalletStorage,
  type Address,
  type Hash,
} from '../src/index';
import { ensureE2EServer, getE2EConfig, shutdownE2EServer } from './setup';

describe('E2E: MPC Wallet Flow', () => {
  let baseURL: string;
  let apiKey: string;
  let client: MPCClient;

  const sdkClientConfig = () => ({
    baseURL,
    apiKey,
    timeout: 120_000,
  });

  beforeAll(async () => {
    await ensureE2EServer();
    const config = getE2EConfig();
    baseURL = config.serverURL;
    apiKey = config.apiKey;
    client = new MPCClient(sdkClientConfig());

    const healthRes = await fetch(baseURL + '/health');
    expect(healthRes.ok).toBe(true);
  }, 60_000);

  afterAll(async () => {
    await shutdownE2EServer();
  });

  it('creates wallet, signs message, and verifies signer address', async () => {
    const wallet = new MPCWallet({ client: sdkClientConfig() });
    const address = await wallet.create(1);

    expect(address).toMatch(/^0x[a-fA-F0-9]{40}$/);

    const message = 'hello-e2e-message';
    const signature = await wallet.signMessage(message);

    const recovered = verifyMessage(message, signature as SignatureLike);
    expect(recovered.toLowerCase()).toBe(address.toLowerCase());
  });

  it('creates wallet, signs tx-shaped payload, and recovers signer from hash signature', async () => {
    const wallet = new MPCWallet({ client: sdkClientConfig() });
    const address = await wallet.create(1);

    const txSignature = await wallet.signTransaction({
      to: '0x742d35cc6634c0532925a3b844bc9e7595f0beb0' as Address,
      value: '1000000000000000',
      gasLimit: '21000',
      nonce: 0,
      chainId: 1,
    });

    expect(txSignature).toMatch(/^0x[a-fA-F0-9]{130}$/);

    const txHash = keccak256(
      toUtf8Bytes('policy-e2e-transaction-hash-anchor-' + address),
    ) as Hash;
    const hashSignature = await wallet.signHash(txHash);

    const recovered = recoverAddress(txHash, hashSignature as SignatureLike);
    expect(recovered.toLowerCase()).toBe(address.toLowerCase());
  });

  it('enforces policy checks for single limit and whitelist', async () => {
    const created = await client.createWallet();
    const address = created.address;

    await client.setPolicy(address, {
      singleTxLimit: '1000',
      whitelist: ['0x742d35cc6634c0532925a3b844bc9e7595f0beb0' as Address],
      dailyTxLimit: 10,
    });

    const messageHash = keccak256(toUtf8Bytes('policy-check'));

    const okRes = await fetch(`${baseURL}/api/v1/wallet/sign`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        address,
        message_hash: messageHash,
        shard1: created.shard1,
        to: '0x742d35cc6634c0532925a3b844bc9e7595f0beb0',
        value: '999',
      }),
    });
    expect(okRes.ok).toBe(true);

    const overLimitRes = await fetch(`${baseURL}/api/v1/wallet/sign`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        address,
        message_hash: messageHash,
        shard1: created.shard1,
        to: '0x742d35cc6634c0532925a3b844bc9e7595f0beb0',
        value: '1001',
      }),
    });
    expect(overLimitRes.ok).toBe(false);

    const nonWhitelistRes = await fetch(`${baseURL}/api/v1/wallet/sign`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        address,
        message_hash: messageHash,
        shard1: created.shard1,
        to: '0x1111111111111111111111111111111111111111',
        value: '1',
      }),
    });
    expect(nonWhitelistRes.ok).toBe(false);
  });

  it('supports multi-wallet creation and repeated signing', async () => {
    const wallets = await Promise.all(
      Array.from({ length: 2 }, async () => {
        const wallet = new MPCWallet({
          client: sdkClientConfig(),
          storage: new MemoryWalletStorage(),
        });
        const address = await wallet.create(1);
        return { wallet, address };
      }),
    );

    expect(new Set(wallets.map((x) => x.address)).size).toBe(2);

    const target = wallets[0];
    const signatures = await Promise.all(
      Array.from({ length: 3 }, (_, i) => target.wallet.signMessage(`concurrent-${i}`)),
    );

    signatures.forEach((sig, i) => {
      const recovered = verifyMessage(`concurrent-${i}`, sig as SignatureLike);
      expect(recovered.toLowerCase()).toBe(target.address.toLowerCase());
    });
  });

  it('persists shard1 with MemoryWalletStorage and can reload', async () => {
    const storage = new MemoryWalletStorage();

    const walletA = new MPCWallet({
      client: sdkClientConfig(),
      storage,
    });

    const address = await walletA.create(1);
    walletA.disconnect();

    const walletB = new MPCWallet({
      client: sdkClientConfig(),
      storage,
    });

    const loaded = await walletB.load(address);
    expect(loaded).toBe(true);

    const sig = await walletB.signMessage('reload-check');
    const recovered = verifyMessage('reload-check', sig as SignatureLike);
    expect(recovered.toLowerCase()).toBe(address.toLowerCase());
  });
});
