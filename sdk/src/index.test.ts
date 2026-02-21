import { describe, expect, it } from 'vitest';

import { MPCClient, MPCError, NetworkError, ValidationError } from './index';

describe('index exports', () => {
  it('should export MPCClient and error classes', () => {
    expect(MPCClient).toBeTypeOf('function');
    expect(MPCError).toBeTypeOf('function');
    expect(NetworkError).toBeTypeOf('function');
    expect(ValidationError).toBeTypeOf('function');
  });
});
