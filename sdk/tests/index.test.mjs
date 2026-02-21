import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const source = readFileSync(resolve(process.cwd(), 'src/index.ts'), 'utf8');

test('SDK exports package name constant', () => {
  assert.match(source, /PACKAGE_NAME\s*=\s*'@agent-vault\/sdk'/);
});

test('SDK exports version constant', () => {
  assert.match(source, /VERSION\s*=\s*'0\.1\.0-alpha'/);
});
