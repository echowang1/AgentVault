# Architecture

## Layers

1. SDK Layer (`sdk/`)
- `MPCClient`: typed HTTP client.
- `MPCWallet`: high-level wallet wrapper, local shard1 handling, events.

2. API Layer (`server/internal/api`)
- Gin routes and middleware.
- API key auth, CORS, unified response.

3. Core Services
- `server/internal/tss`: key generation + signing (GG18, secp256k1).
- `server/internal/policy`: tx policy checks and usage accounting.
- `server/internal/storage`: SQLite persistence, AES-256-GCM shard encryption.

## Key Data Flow

Create wallet:
1. Client calls `POST /api/v1/wallet/create`.
2. Server runs TSS keygen and generates address/public key.
3. `shard1` returns to caller, `shard2` encrypted and stored server-side.

Sign:
1. Client submits `address`, `message_hash`, `shard1`.
2. Server loads `shard2`, checks policy (`to`, `value` if provided).
3. TSS signing returns ECDSA signature `(r,s,v)`.

## Security Boundaries

- Shard split model: `shard1` client-side, `shard2` server-side.
- `shard2` is encrypted at rest using `SHARD_ENCRYPTION_KEY`.
- API endpoints require Bearer API key except `/health`.
