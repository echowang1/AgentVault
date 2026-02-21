import { describe, expect, it } from 'vitest';

import {
  MPCClient,
  MPCError,
  MPCWallet,
  MemoryWalletStorage,
  NetworkError,
  ValidationError,
} from './index';

describe('index exports', () => {
  it('should export SDK classes and errors', () => {
    expect(MPCClient).toBeTypeOf('function');
    expect(MPCWallet).toBeTypeOf('function');
    expect(MemoryWalletStorage).toBeTypeOf('function');
    expect(MPCError).toBeTypeOf('function');
    expect(NetworkError).toBeTypeOf('function');
    expect(ValidationError).toBeTypeOf('function');
  });
});
