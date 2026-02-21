import { spawn, type ChildProcessWithoutNullStreams } from 'node:child_process';
import path from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';

const SERVER_URL = process.env.E2E_SERVER_URL ?? 'http://127.0.0.1:18080';
const API_KEY = process.env.E2E_API_KEY ?? 'test-api-key-e2e';
const ENCRYPTION_KEY_B64 =
  process.env.SHARD_ENCRYPTION_KEY ?? Buffer.alloc(32, 7).toString('base64');

let serverProcess: ChildProcessWithoutNullStreams | undefined;
let startedByTest = false;

async function waitForHealth(url: string, timeoutMs: number): Promise<void> {
  const started = Date.now();

  while (Date.now() - started < timeoutMs) {
    try {
      const res = await fetch(`${url}/health`);
      if (res.ok) {
        return;
      }
    } catch {
      // server not ready
    }

    await delay(250);
  }

  throw new Error(`E2E server is not healthy after ${timeoutMs}ms: ${url}`);
}

export async function ensureE2EServer(): Promise<void> {
  if (serverProcess || process.env.E2E_SERVER_MANAGED === 'external') {
    await waitForHealth(SERVER_URL, 8000);
    return;
  }

  const repoRoot = path.resolve(process.cwd(), '..');

  serverProcess = spawn('go', ['run', './cmd/server/main.go'], {
    cwd: path.join(repoRoot, 'server'),
    env: {
      ...process.env,
      SERVER_HOST: '127.0.0.1',
      SERVER_PORT: '18080',
      MPC_API_KEYS: API_KEY,
      DB_PATH: 'file:agentvault_e2e?mode=memory&cache=shared',
      SHARD_ENCRYPTION_KEY: ENCRYPTION_KEY_B64,
    },
    stdio: 'pipe',
  });

  startedByTest = true;

  serverProcess.stdout.on('data', (chunk) => {
    process.stdout.write(`[e2e-server] ${chunk}`);
  });

  serverProcess.stderr.on('data', (chunk) => {
    process.stderr.write(`[e2e-server] ${chunk}`);
  });

  // CI cold-start for `go run` can be slow on first compile.
  await waitForHealth(SERVER_URL, 90000);
}

export async function shutdownE2EServer(): Promise<void> {
  if (!serverProcess || !startedByTest) {
    return;
  }

  await new Promise<void>((resolve) => {
    const proc = serverProcess;
    serverProcess = undefined;

    proc.once('exit', () => resolve());
    proc.kill('SIGTERM');

    setTimeout(() => {
      if (!proc.killed) {
        proc.kill('SIGKILL');
      }
      resolve();
    }, 3000);
  });
}

export function getE2EConfig(): { serverURL: string; apiKey: string } {
  return {
    serverURL: SERVER_URL,
    apiKey: API_KEY,
  };
}
