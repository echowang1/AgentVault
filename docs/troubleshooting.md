# Troubleshooting

## Server fails on startup

Symptom:
- `SHARD_ENCRYPTION_KEY is required`

Fix:
- Set `SHARD_ENCRYPTION_KEY` to base64 32-byte key.

## 401 unauthorized

Symptom:
- API returns `UNAUTHORIZED`.

Fix:
- Ensure request header includes `Authorization: Bearer <key>`.
- Check key exists in `MPC_API_KEYS`.

## Wallet sign fails with invalid hash

Symptom:
- API returns validation error.

Fix:
- `message_hash` must match `0x` + 64 hex chars.

## Policy rejection

Symptom:
- `POST /wallet/sign` returns `INVALID_REQUEST` with policy message.

Fix:
- Inspect policy (`GET /wallet/:address/policy`).
- Check `to`/`value` and time window constraints.

## SQLite runtime issues

Symptom:
- DB open/migration errors.

Fix:
- Verify `DB_PATH` is writable.
- Ensure mounted volume permissions are correct in Docker.

## SDK timeout under heavy load

Symptom:
- `NetworkError: request timed out`

Fix:
- Increase SDK timeout, e.g. `timeout: 120000`.
- Reduce concurrent keygen/sign operations in one batch.
