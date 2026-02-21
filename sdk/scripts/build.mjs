import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

const distDir = resolve(process.cwd(), 'dist');
mkdirSync(distDir, { recursive: true });

writeFileSync(
  resolve(distDir, 'index.js'),
  "'use strict';\n\n" +
    "const PACKAGE_NAME = '@agent-vault/sdk';\n" +
    "const VERSION = '0.1.0-alpha';\n\n" +
    'module.exports = { PACKAGE_NAME, VERSION };\n',
);

writeFileSync(
  resolve(distDir, 'index.d.ts'),
  "export declare const PACKAGE_NAME = '@agent-vault/sdk';\n" +
    "export declare const VERSION = '0.1.0-alpha';\n" +
    'export interface WalletConfig {\n' +
    '  baseURL: string;\n' +
    '  apiKey: string;\n' +
    '}\n' +
    'export interface SignRequest {\n' +
    '  address: string;\n' +
    '  messageHash: string;\n' +
    '  shard1: string;\n' +
    '}\n' +
    'export interface SignResponse {\n' +
    '  signature: string;\n' +
    '  r: string;\n' +
    '  s: string;\n' +
    '  v: number;\n' +
    '}\n',
);

console.log('SDK build completed.');
