# AgentVault SDK Guide

## Install

```bash
cd sdk
npm install
npm run build
```

## MPCClient

```ts
import { MPCClient } from './dist/index.js';

const client = new MPCClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'test-api-key',
});

const wallet = await client.createWallet();
const sig = await client.sign({
  address: wallet.address,
  messageHash: '0x' + '11'.repeat(32),
  shard1: wallet.shard1,
});
```

Methods:
- `createWallet(req?)`
- `sign(req)`
- `getWallet(address)`
- `setPolicy(address, policy)`
- `getPolicy(address)`
- `getDailyUsage(address)`
- `healthCheck()`

## MPCWallet

```ts
import { MPCWallet, MemoryWalletStorage } from './dist/index.js';

const wallet = new MPCWallet({
  client: {
    baseURL: 'http://localhost:8080',
    apiKey: 'test-api-key',
    timeout: 120000,
  },
  storage: new MemoryWalletStorage(),
});

const address = await wallet.create(1);
const signature = await wallet.signMessage('hello');
```

Main methods:
- `create(chainId?)`
- `connect(address, shard1)`
- `load(address)`
- `disconnect()`
- `getAddress()`
- `signTransaction(tx, options?)`
- `signMessage(message, options?)`
- `signHash(hash, options?)`
- `setPolicy(policy)`
- `getPolicy()`
- `getDailyUsage()`
- `on(event, listener)` / `off(event, listener)`

Event types:
- `beforeSign`
- `afterSign`
- `policyCheck`
- `error`

## Storage Options

- `MemoryWalletStorage`: in-memory, useful for tests.
- `LocalStorageWalletStorage`: browser localStorage persistence.
