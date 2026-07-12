/**
 * Lazy-load secrets from SSM Parameter Store (SecureString).
 * Cached in-process so warm Lambdas do not hit SSM on every chat turn.
 *
 * Local override: set ANTHROPIC_API_KEY in .env.local (never commit).
 * Production: only the parameter *name* is in Lambda env — not the key value.
 */

let cachedAnthropicKey: string | null = null;
let cachedAnthropicAt = 0;

const CACHE_TTL_MS = 10 * 60 * 1000; // 10 minutes

const DEFAULT_SSM_PATH = '/pfrs-lab/production/anthropic-api-key';

export function anthropicApiKeySsmPath(): string {
  return process.env.ANTHROPIC_API_KEY_SSM || DEFAULT_SSM_PATH;
}

export async function getAnthropicApiKey(): Promise<string> {
  const fromEnv = process.env.ANTHROPIC_API_KEY?.trim();
  if (fromEnv) {
    return fromEnv;
  }

  const now = Date.now();
  if (cachedAnthropicKey && now - cachedAnthropicAt < CACHE_TTL_MS) {
    return cachedAnthropicKey;
  }

  const name = anthropicApiKeySsmPath();
  // Dynamic require keeps client out of cold paths that never chat.
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const { SSMClient, GetParameterCommand } = require('@aws-sdk/client-ssm');

  const client = new SSMClient({});
  const out = await client.send(
    new GetParameterCommand({
      Name: name,
      WithDecryption: true,
    }),
  );

  const value = out.Parameter?.Value?.trim();
  if (!value) {
    throw new Error(
      `SSM parameter ${name} is empty or missing. Put a SecureString there (CI sync or aws ssm put-parameter).`,
    );
  }

  cachedAnthropicKey = value;
  cachedAnthropicAt = now;
  return value;
}
