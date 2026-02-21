import { MPCWallet } from '../sdk.js';

type RuntimeLike = {
  getSetting(key: string): string | undefined;
  registerAction(action: {
    name: string;
    description: string;
    handler: (args: Record<string, unknown>) => Promise<Record<string, unknown>>;
  }): void;
};

export class MPCWalletPlugin {
  name = 'mpc-wallet';
  description = 'AgentVault MPC wallet plugin example for ElizaOS';

  private wallet?: MPCWallet;

  async init(runtime: RuntimeLike): Promise<void> {
    const baseURL = runtime.getSetting('MPC_SERVER_URL');
    const apiKey = runtime.getSetting('MPC_API_KEY');

    if (!baseURL || !apiKey) {
      throw new Error('Missing MPC_SERVER_URL or MPC_API_KEY');
    }

    this.wallet = new MPCWallet({
      client: { baseURL, apiKey },
    });

    const address = runtime.getSetting('MPC_WALLET_ADDRESS') as `0x${string}` | undefined;
    const shard1 = runtime.getSetting('MPC_WALLET_SHARD1');

    if (address && shard1) {
      await this.wallet.connect(address, shard1);
      console.log(`[mpc-wallet] connected wallet ${address}`);
    } else {
      const newAddress = await this.wallet.create(1);
      console.log(`[mpc-wallet] created new wallet ${newAddress}`);
      console.log('[mpc-wallet] configure storage to persist shard1 in production.');
    }

    runtime.registerAction({
      name: 'SIGN_MESSAGE',
      description: 'Sign message with MPC wallet',
      handler: async (args) => {
        if (!this.wallet) {
          throw new Error('Wallet is not initialized');
        }
        const message = String(args.message ?? '');
        const signature = await this.wallet.signMessage(message);
        return { success: true, signature };
      },
    });

    runtime.registerAction({
      name: 'SIGN_TRANSACTION',
      description: 'Sign transaction with MPC wallet',
      handler: async (args) => {
        if (!this.wallet) {
          throw new Error('Wallet is not initialized');
        }

        const signature = await this.wallet.signTransaction({
          to: args.to as `0x${string}`,
          value: args.value as string,
          data: args.data as string | undefined,
          gasLimit: args.gasLimit as string | undefined,
          nonce: args.nonce as number | undefined,
          chainId: args.chainId as number | undefined,
        });

        return { success: true, signature };
      },
    });
  }

  async cleanup(): Promise<void> {
    this.wallet?.disconnect();
  }
}

export default MPCWalletPlugin;
