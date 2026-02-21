import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import { MPCWallet } from '../src/index';
import { ensureE2EServer, getE2EConfig, shutdownE2EServer } from './setup';

describe('Performance: MPC Wallet', () => {
  let baseURL: string;
  let apiKey: string;

  beforeAll(async () => {
    await ensureE2EServer();
    const config = getE2EConfig();
    baseURL = config.serverURL;
    apiKey = config.apiKey;
  }, 60_000);

  afterAll(async () => {
    await shutdownE2EServer();
  });

  it('creates wallet within threshold', async () => {
    const start = performance.now();
    const wallet = new MPCWallet({ client: { baseURL, apiKey, timeout: 120_000 } });
    await wallet.create(1);
    const elapsed = performance.now() - start;

    expect(elapsed).toBeLessThan(90_000);
    console.log(`create wallet: ${elapsed.toFixed(1)}ms`);
  }, 120_000);

  it('signs one message within threshold', async () => {
    const wallet = new MPCWallet({ client: { baseURL, apiKey, timeout: 120_000 } });
    await wallet.create(1);

    const start = performance.now();
    await wallet.signMessage('perf-sign-single');
    const elapsed = performance.now() - start;

    expect(elapsed).toBeLessThan(30_000);
    console.log(`sign single message: ${elapsed.toFixed(1)}ms`);
  }, 120_000);

  it('handles batch signing throughput', async () => {
    const wallet = new MPCWallet({ client: { baseURL, apiKey, timeout: 120_000 } });
    await wallet.create(1);

    const count = 5;
    const start = performance.now();
    const signatures = await Promise.all(
      Array.from({ length: count }, (_, i) => wallet.signMessage(`perf-batch-${i}`)),
    );
    const elapsed = performance.now() - start;

    expect(signatures).toHaveLength(count);
    expect(elapsed).toBeLessThan(120_000);
    console.log(`batch sign ${count}: ${elapsed.toFixed(1)}ms`);
  }, 180_000);
});
