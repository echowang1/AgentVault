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
    console.log('== AgentVault Example: Set Policy ==');
    await wallet.connect(EXAMPLE_ADDRESS, EXAMPLE_SHARD1);

    await wallet.setPolicy({
      singleTxLimit: parseEther('1').toString(),
      dailyLimit: parseEther('5').toString(),
      whitelist: [
        '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb',
        '0xE592427A0AEce92De3Edee1F18E0157C05861564',
      ],
      dailyTxLimit: 20,
      startTime: '00:00',
      endTime: '23:59',
    });

    console.log('Policy updated successfully');

    const policy = await wallet.getPolicy();
    const usage = await wallet.getDailyUsage();

    console.log('Current policy:');
    console.log(JSON.stringify(policy, null, 2));
    console.log('Today usage:');
    console.log(JSON.stringify(usage, null, 2));
  } catch (error) {
    console.error('Failed to set or query policy');
    console.error(error);
    process.exitCode = 1;
  }
}

void main();
