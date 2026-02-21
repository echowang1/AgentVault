# GOAT SDK Adapter Example

`mpc-wallet-adapter.ts` shows how to adapt `MPCWallet` to a GOAT-like EVM wallet client.

## What it includes

- `getAddress`
- `signTransaction`
- `signMessage`
- `signTypedData` (using ethers `TypedDataEncoder.hash` + `wallet.signHash`)

## Notes

- Interface is intentionally lightweight to keep the example framework-agnostic.
- Replace local type aliases with official GOAT interfaces in your app.
