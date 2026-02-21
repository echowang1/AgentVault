import { MemoryWalletStorage, MPCWallet } from '../sdk.js';

async function main(): Promise<void> {
  const baseURL = process.env.MPC_SERVER_URL ?? 'http://localhost:8080';
  const apiKey = process.env.MPC_API_KEY ?? 'test-api-key';

  const wallet = new MPCWallet({
    client: { baseURL, apiKey },
    storage: new MemoryWalletStorage(),
  });

  try {
    console.log('== AgentVault Example: Create Wallet ==');
    console.log(`MPC server: ${baseURL}`);

    const address = await wallet.create(1);

    console.log('Wallet created successfully');
    console.log(`Address: ${address}`);
    console.log('Shard1 is managed by wallet storage; keep your storage secure.');
  } catch (error) {
    console.error('Failed to create wallet');
    console.error(error);
    process.exitCode = 1;
  }
}

void main();
