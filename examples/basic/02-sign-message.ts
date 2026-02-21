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
    console.log('== AgentVault Example: Sign Message ==');
    await wallet.connect(EXAMPLE_ADDRESS, EXAMPLE_SHARD1);

    const message = process.env.MPC_MESSAGE ?? 'Hello from AgentVault SDK';
    console.log(`Message: ${message}`);

    const signature = await wallet.signMessage(message);

    console.log('Message signed successfully');
    console.log(`Signature: ${signature}`);
  } catch (error) {
    console.error('Failed to sign message');
    console.error(error);
    process.exitCode = 1;
  }
}

void main();
