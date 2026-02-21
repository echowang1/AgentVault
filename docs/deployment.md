# Deployment Guide

## Environment Variables

Required:
- `MPC_API_KEYS`: comma-separated API keys.
- `SHARD_ENCRYPTION_KEY`: base64-encoded 32-byte key.

Optional:
- `SERVER_HOST` (default `0.0.0.0`)
- `SERVER_PORT` (default `8080`)
- `DB_PATH` (default `./data/mpc-wallet.db`)

Generate encryption key:

```bash
openssl rand -base64 32
```

## Docker

Build:

```bash
docker build -t agent-vault:latest -f docker/Dockerfile .
```

Run:

```bash
docker run -d \
  --name agent-vault \
  -p 8080:8080 \
  -e MPC_API_KEYS=test-api-key \
  -e SHARD_ENCRYPTION_KEY=<base64-32-byte-key> \
  -e DB_PATH=/app/data/mpc-wallet.db \
  -v $(pwd)/data:/app/data \
  agent-vault:latest
```

## Docker Compose

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

If using `docker/docker-compose.yml`, add `SHARD_ENCRYPTION_KEY` to environment before start.

## Kubernetes (Reference)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-vault
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agent-vault
  template:
    metadata:
      labels:
        app: agent-vault
    spec:
      containers:
        - name: agent-vault
          image: ghcr.io/echowang1/agent-vault:latest
          ports:
            - containerPort: 8080
          env:
            - name: MPC_API_KEYS
              valueFrom:
                secretKeyRef:
                  name: agent-vault-secret
                  key: api-keys
            - name: SHARD_ENCRYPTION_KEY
              valueFrom:
                secretKeyRef:
                  name: agent-vault-secret
                  key: shard-encryption-key
```

## Smoke Check

```bash
curl http://localhost:8080/health
```
