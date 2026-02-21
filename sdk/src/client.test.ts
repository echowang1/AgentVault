import { beforeEach, describe, expect, it, vi } from 'vitest';

import { MPCClient } from './client';
import { AuthError, MPCError, ValidationError } from './errors';

describe('MPCClient', () => {
  let mockFetch: ReturnType<typeof vi.fn>;
  let client: MPCClient;

  beforeEach(() => {
    mockFetch = vi.fn();
    client = new MPCClient({
      baseURL: 'http://localhost:8080',
      apiKey: 'test-api-key',
      fetch: mockFetch as unknown as typeof fetch,
      timeout: 500,
    });
  });

  it('createWallet should send request and map response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        success: true,
        data: {
          address: '0x1234567890123456789012345678901234567890',
          public_key: '0xabc',
          shard1: 'base64-shard',
          shard2_id: 'shard-2',
        },
      }),
    });

    const result = await client.createWallet();

    expect(result).toEqual({
      address: '0x1234567890123456789012345678901234567890',
      publicKey: '0xabc',
      shard1: 'base64-shard',
      shard2Id: 'shard-2',
    });

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/wallet/create',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          Authorization: 'Bearer test-api-key',
        }),
      }),
    );
  });

  it('sign should validate address', async () => {
    await expect(
      client.sign({
        address: 'invalid' as never,
        messageHash: '0x' + 'a'.repeat(64) as never,
        shard1: 'test',
      }),
    ).rejects.toBeInstanceOf(ValidationError);
  });

  it('sign should validate hash', async () => {
    await expect(
      client.sign({
        address: '0x1234567890123456789012345678901234567890',
        messageHash: 'invalid' as never,
        shard1: 'test',
      }),
    ).rejects.toBeInstanceOf(ValidationError);
  });

  it('should throw MPCError for API error response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      statusText: 'Bad Request',
      json: async () => ({
        success: false,
        error: {
          code: 'INVALID_REQUEST',
          message: 'bad input',
        },
      }),
    });

    await expect(client.createWallet()).rejects.toEqual(
      expect.objectContaining({
        name: 'MPCError',
        code: 'INVALID_REQUEST',
        message: 'bad input',
      }),
    );
  });

  it('should throw AuthError on 401', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      json: async () => ({
        success: false,
        error: {
          code: 'UNAUTHORIZED',
          message: 'invalid api key',
        },
      }),
    });

    await expect(client.createWallet()).rejects.toBeInstanceOf(AuthError);
  });

  it('healthCheck should work with plain response shape', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true, data: { status: 'ok', version: '0.1.0' } }),
    });

    const result = await client.healthCheck();
    expect(result).toEqual({ status: 'ok', version: '0.1.0' });
  });

  it('network failure should map to Error/MPC hierarchy', async () => {
    mockFetch.mockRejectedValueOnce(new TypeError('network down'));

    await expect(client.createWallet()).rejects.toEqual(
      expect.objectContaining({ name: 'NetworkError' }),
    );
  });

  it('timeout should map to NetworkError', async () => {
    mockFetch.mockImplementationOnce(async (_url, init) => {
      if (init?.signal) {
        await new Promise((resolve, reject) => {
          init.signal?.addEventListener('abort', () => reject(new DOMException('aborted', 'AbortError')));
          setTimeout(resolve, 1000);
        });
      }
      return {
        ok: true,
        json: async () => ({ success: true, data: {} }),
      };
    });

    await expect(client.createWallet()).rejects.toEqual(
      expect.objectContaining({ name: 'NetworkError' }),
    );
  });

  it('MPCError class should keep details', () => {
    const err = new MPCError('X', 'msg', { a: 1 });
    expect(err.details).toEqual({ a: 1 });
  });
});
