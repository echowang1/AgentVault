# ElizaOS Plugin Example

`mpc-wallet-plugin.ts` demonstrates how to wrap `MPCWallet` as an ElizaOS-style plugin.

## What it shows

- Read `MPC_SERVER_URL` and `MPC_API_KEY` from runtime settings
- Connect existing wallet (`MPC_WALLET_ADDRESS` + `MPC_WALLET_SHARD1`) or create one
- Register `SIGN_MESSAGE` and `SIGN_TRANSACTION` actions

## Notes

- This is an integration template, not a full production plugin package.
- Add secure shard1 persistence before production use.
