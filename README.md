# AgentVault

[![Go Tests](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/go-test.yml/badge.svg)](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/go-test.yml)
[![TS Tests](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/ts-test.yml/badge.svg)](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/ts-test.yml)
[![E2E Tests](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/e2e.yml/badge.svg)](https://github.com/echowang1/agent-mpc-wallet/actions/workflows/e2e.yml)

MPC wallet service and TypeScript SDK for AI agents.

## Highlights

- 2-of-2 MPC (GG18) key generation and signing.
- HTTP API with API key auth.
- Policy engine (limits, whitelist, tx count, time window).
- SQLite persistence with AES-256-GCM encrypted shard storage.
- TypeScript SDK (`MPCClient`, `MPCWallet`) and examples.

## Quick Start

### 1) Run server (Docker)

```bash
docker build -t agent-vault:latest -f docker/Dockerfile .
docker run -d \
  --name agent-vault \
  -p 8080:8080 \
  -e MPC_API_KEYS=test-api-key \
  -e SHARD_ENCRYPTION_KEY=$(openssl rand -base64 32) \
  -v $(pwd)/data:/app/data \
  agent-vault:latest
```

### 2) Build SDK

```bash
cd sdk
npm install
npm run build
```

### 3) Use SDK

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
const signature = await wallet.signMessage('Hello AgentVault');

console.log(address, signature);
```

## API Endpoints

- `GET /health`
- `POST /api/v1/wallet/create`
- `POST /api/v1/wallet/sign`
- `GET /api/v1/wallet/:address`
- `PUT /api/v1/wallet/:address/policy`
- `GET /api/v1/wallet/:address/policy`
- `GET /api/v1/wallet/:address/usage`

## Documentation

- `docs/api.md`
- `docs/sdk.md`
- `docs/deployment.md`
- `docs/architecture.md`
- `docs/security.md`
- `docs/troubleshooting.md`
- `docs/openapi.json`

## Examples

- `examples/basic/README.md`
- `examples/with-eliza/README.md`
- `examples/with-goat/README.md`

## License

MIT. See `LICENSE`.
