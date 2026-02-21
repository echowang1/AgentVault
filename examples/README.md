# AgentVault Examples

This directory contains usage examples for AgentVault SDK.

## Environment Variables

```bash
export MPC_SERVER_URL=http://localhost:8080
export MPC_API_KEY=your-api-key
# for sign/policy examples
export MPC_WALLET_ADDRESS=0x...
export MPC_WALLET_SHARD1=...
```

## Install and Build

```bash
cd sdk
npm install
npm run build
```

## Basic Examples

Run from `sdk/`:

```bash
npm run example:create
npm run example:sign-message
npm run example:sign-tx
npm run example:set-policy
```

## Integrations

- ElizaOS plugin template: `examples/with-eliza/mpc-wallet-plugin.ts`
- GOAT adapter template: `examples/with-goat/mpc-wallet-adapter.ts`
