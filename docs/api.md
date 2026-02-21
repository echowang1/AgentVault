# AgentVault API Documentation

## Base

- Base URL: `http://localhost:8080`
- Auth: `Authorization: Bearer <API_KEY>`
- Content-Type: `application/json`

## Response Format

Most API endpoints return:

```json
{
  "success": true,
  "data": {}
}
```

Error format:

```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "...",
    "details": {}
  }
}
```

## Endpoints

### GET /health

Health endpoint (no API key required).

Response:

```json
{
  "status": "ok",
  "version": "0.1.0",
  "timestamp": "2026-02-21T18:00:00Z"
}
```

### POST /api/v1/wallet/create

Create MPC wallet.

Request body (optional):

```json
{
  "chain_id": 1
}
```

Response `data`:

```json
{
  "address": "0x...",
  "public_key": "04...",
  "shard1": "base64...",
  "shard2_id": "..."
}
```

### POST /api/v1/wallet/sign

Sign 32-byte message hash with MPC shards.

Request body:

```json
{
  "address": "0x...",
  "message_hash": "0x<64-hex>",
  "shard1": "base64...",
  "shard2_id": "optional",
  "to": "0x...",
  "value": "1000"
}
```

Notes:
- `to` and `value` are used by policy engine checks.
- `message_hash` must be exactly 32 bytes hex.

Response `data`:

```json
{
  "signature": "0x<rsv>",
  "r": "0x...",
  "s": "0x...",
  "v": 27
}
```

### GET /api/v1/wallet/:address

Get wallet info.

Response `data`:

```json
{
  "address": "0x...",
  "public_key": "04...",
  "created_at": "2026-02-21T18:00:00Z"
}
```

### PUT /api/v1/wallet/:address/policy

Set wallet policy.

Request body:

```json
{
  "single_tx_limit": "1000000000000000000",
  "daily_limit": "10000000000000000000",
  "whitelist": ["0x..."],
  "daily_tx_limit": 100,
  "start_time": "09:00",
  "end_time": "18:00"
}
```

Response `data`:

```json
{
  "wallet_id": "..."
}
```

### GET /api/v1/wallet/:address/policy

Get current wallet policy.

### GET /api/v1/wallet/:address/usage

Get current daily usage.

Response `data`:

```json
{
  "wallet_id": "...",
  "date": "2026-02-21",
  "total_amount": "123456",
  "tx_count": 2
}
```

## Error Codes

Current API-level error codes:

- `UNAUTHORIZED`
- `INVALID_REQUEST`
- `NOT_FOUND`
- `INTERNAL_ERROR`

Policy violations are returned as `INVALID_REQUEST` with descriptive `message`.
