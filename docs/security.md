# Security Best Practices

## Secrets

- Store `MPC_API_KEYS` and `SHARD_ENCRYPTION_KEY` in secret manager/K8s Secret.
- Rotate API keys and encryption keys periodically.
- Never commit real shard data or secret env files.

## Shard Handling

- Keep `shard1` only in secure client storage.
- Do not print shards in logs.
- Use dedicated storage encryption in browser/mobile if persisted.

## Transport and Network

- Serve API behind HTTPS/TLS.
- Restrict server access via firewall/private network.
- Use rate limiting and WAF in production ingress.

## Runtime Hardening

- Run container with non-root user where possible.
- Pin image tags and scan vulnerabilities.
- Enable centralized logging and audit trails.

## Policy

- Enforce `single_tx_limit`, `daily_limit`, whitelist, and time windows.
- Keep whitelist addresses normalized (lowercase checksummed consistency in apps).
