# Basic Examples

These examples show the minimal workflow for AgentVault SDK.

## Prerequisites

1. Build SDK first:

```bash
cd sdk
npm run build
```

2. Set env vars:

```bash
export MPC_SERVER_URL=http://localhost:8080
export MPC_API_KEY=your-api-key
export MPC_WALLET_ADDRESS=0x...
export MPC_WALLET_SHARD1=...
```

## Run

From `sdk/` directory:

```bash
npm run example:create
npm run example:sign-message
npm run example:sign-tx
npm run example:set-policy
```

`01-create-wallet.ts` can run without pre-existing wallet. Other scripts require `MPC_WALLET_ADDRESS` and `MPC_WALLET_SHARD1`.
