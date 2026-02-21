import { beforeEach, describe, expect, it, vi } from 'vitest';

import { MemoryWalletStorage } from './storage/memory';
import type { Address, Hash } from './types';
import { MPCWallet } from './wallet';

describe('MPCWallet', () => {
  const mockAddress = '0x1234567890123456789012345678901234567890' as Address;

  let wallet: MPCWallet;
  let mockClient: {
    createWallet: ReturnType<typeof vi.fn>;
    getWallet: ReturnType<typeof vi.fn>;
    sign: ReturnType<typeof vi.fn>;
    setPolicy: ReturnType<typeof vi.fn>;
    getPolicy: ReturnType<typeof vi.fn>;
    getDailyUsage: ReturnType<typeof vi.fn>;
  };

  beforeEach(() => {
    mockClient = {
      createWallet: vi.fn(),
      getWallet: vi.fn(),
      sign: vi.fn(),
      setPolicy: vi.fn(),
      getPolicy: vi.fn(),
      getDailyUsage: vi.fn(),
    };

    wallet = new MPCWallet({
      client: { baseURL: 'http://localhost:8080', apiKey: 'test-api-key' },
      storage: new MemoryWalletStorage(),
    });

    (wallet as unknown as { client: typeof mockClient }).client = mockClient;
  });

  it('create should cache and persist wallet data', async () => {
    mockClient.createWallet.mockResolvedValueOnce({
      address: mockAddress,
      publicKey: '0xpub',
      shard1: 'shard-1',
      shard2Id: 'shard-2',
    });

    const created = await wallet.create(1);
    expect(created).toBe(mockAddress);
    expect(wallet.getAddress()).toBe(mockAddress);
    expect(wallet.getChainId()).toBe(1);
    expect(mockClient.createWallet).toHaveBeenCalledWith({ chainId: 1 });

    const loaded = await wallet.load(mockAddress);
    expect(loaded).toBe(true);
  });

  it('connect should verify wallet and persist shard', async () => {
    mockClient.getWallet.mockResolvedValueOnce({
      address: mockAddress,
      publicKey: '0xpub',
      createdAt: '2026-01-01T00:00:00Z',
    });

    await wallet.connect(mockAddress, 'agent-shard');
    expect(wallet.getAddress()).toBe(mockAddress);
    expect(mockClient.getWallet).toHaveBeenCalledWith(mockAddress);
  });

  it('load should return false when storage has no data', async () => {
    const found = await wallet.load(mockAddress);
    expect(found).toBe(false);
  });

  it('signHash should call client.sign with current shard', async () => {
    mockClient.createWallet.mockResolvedValueOnce({
      address: mockAddress,
      publicKey: '0xpub',
      shard1: 'shard-1',
      shard2Id: 'shard-2',
    });
    mockClient.sign.mockResolvedValueOnce({
      signature: ('0x' + 'ab'.repeat(65)) as Hash,
      r: '0x' + '11'.repeat(32),
      s: '0x' + '22'.repeat(32),
      v: 27,
    });
    await wallet.create();

    const hash = ('0x' + '01'.repeat(32)) as Hash;
    const signature = await wallet.signHash(hash);

    expect(signature).toBe('0x' + 'ab'.repeat(65));
    expect(mockClient.sign).toHaveBeenCalledWith({
      address: mockAddress,
      messageHash: hash,
      shard1: 'shard-1',
    });
  });

  it('signMessage should hash and sign with event hooks', async () => {
    const beforeListener = vi.fn();
    const afterListener = vi.fn();
    const policyListener = vi.fn();

    wallet.on('beforeSign', beforeListener);
    wallet.on('afterSign', afterListener);
    wallet.on('policyCheck', policyListener);

    mockClient.createWallet.mockResolvedValueOnce({
      address: mockAddress,
      publicKey: '0xpub',
      shard1: 'shard-1',
      shard2Id: 'shard-2',
    });
    mockClient.sign.mockResolvedValueOnce({
      signature: ('0x' + 'cd'.repeat(65)) as Hash,
      r: '0x' + '11'.repeat(32),
      s: '0x' + '22'.repeat(32),
      v: 28,
    });
    await wallet.create();

    const signature = await wallet.signMessage('hello');
    expect(signature).toBe('0x' + 'cd'.repeat(65));
    expect(policyListener).toHaveBeenCalledTimes(1);
    expect(beforeListener).toHaveBeenCalledTimes(1);
    expect(afterListener).toHaveBeenCalledTimes(1);
  });

  it('should emit error event when sign fails', async () => {
    const errorListener = vi.fn();
    wallet.on('error', errorListener);

    mockClient.createWallet.mockResolvedValueOnce({
      address: mockAddress,
      publicKey: '0xpub',
      shard1: 'shard-1',
      shard2Id: 'shard-2',
    });
    mockClient.sign.mockRejectedValueOnce(new Error('sign failed'));
    await wallet.create();

    await expect(wallet.signHash(('0x' + '01'.repeat(32)) as Hash)).rejects.toThrow('sign failed');
    expect(errorListener).toHaveBeenCalledTimes(1);
  });

  it('disconnect should clear in-memory wallet state', async () => {
    mockClient.createWallet.mockResolvedValueOnce({
      address: mockAddress,
      publicKey: '0xpub',
      shard1: 'shard-1',
      shard2Id: 'shard-2',
    });
    await wallet.create();

    wallet.disconnect();
    expect(wallet.getAddress()).toBeUndefined();
    await expect(wallet.signMessage('test')).rejects.toThrow('Wallet not connected');
  });

  it('policy methods should proxy MPCClient methods', async () => {
    mockClient.createWallet.mockResolvedValueOnce({
      address: mockAddress,
      publicKey: '0xpub',
      shard1: 'shard-1',
      shard2Id: 'shard-2',
    });
    mockClient.getPolicy.mockResolvedValueOnce({
      walletId: 'w1',
      dailyLimit: '1000',
    });
    mockClient.getDailyUsage.mockResolvedValueOnce({
      date: '2026-02-21',
      totalAmount: '1',
      txCount: 1,
    });

    await wallet.create();
    await wallet.setPolicy({ dailyLimit: '1000' });
    const policy = await wallet.getPolicy();
    const usage = await wallet.getDailyUsage();

    expect(mockClient.setPolicy).toHaveBeenCalledWith(mockAddress, { dailyLimit: '1000' });
    expect(policy?.walletId).toBe('w1');
    expect(usage?.txCount).toBe(1);
  });
});
