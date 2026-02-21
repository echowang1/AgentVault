import { parseEther } from 'ethers';

import { MPCWallet } from '../sdk.js';

const EXAMPLE_ADDRESS =
  (process.env.MPC_WALLET_ADDRESS as `0x${string}` | undefined) ??
  '0x1234567890123456789012345678901234567890';
const EXAMPLE_SHARD1 = process.env.MPC_WALLET_SHARD1 ?? 'replace-with-shard1';

async function main(): Promise<void> {
  const wallet = new MPCWallet({
    client: {
      baseURL: process.env.MPC_SERVER_URL ?? 'http://localhost:8080',
      apiKey: process.env.MPC_API_KEY ?? 'test-api-key',
    },
  });

  try {
    console.log('== AgentVault Example: Sign Transaction ==');
    await wallet.connect(EXAMPLE_ADDRESS, EXAMPLE_SHARD1);

    const tx = {
      to: '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb' as const,
      value: parseEther('0.001').toString(),
      gasLimit: '21000',
      chainId: 1,
      nonce: 0,
    };

    console.log('Transaction payload:');
    console.log(JSON.stringify(tx, null, 2));

    const signature = await wallet.signTransaction(tx);

    console.log('Transaction signed successfully');
    console.log(`Signature: ${signature}`);
  } catch (error) {
    console.error('Failed to sign transaction');
    console.error(error);
    process.exitCode = 1;
  }
}

void main();
