import { describe, expect, it } from 'vitest';
import { PACKAGE_NAME, VERSION } from './index';

describe('SDK', () => {
  it('should export package name', () => {
    expect(PACKAGE_NAME).toBe('@agent-vault/sdk');
  });

  it('should export version', () => {
    expect(VERSION).toBe('0.1.0-alpha');
  });
});
