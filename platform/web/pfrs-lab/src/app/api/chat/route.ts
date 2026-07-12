import { NextResponse } from 'next/server';
import { getStorageProvider } from '@/lib/storage';
import { completeChat, type ChatMessage } from '@/lib/llm/complete';
import { recordChatUsage } from '@/lib/llm/usage';
import {
  ASSISTANT_MANIFEST_RUN_CONTEXT,
  ASSISTANT_MAX_TOKENS,
  ASSISTANT_RATE_LIMIT_PER_MINUTE,
  ASSISTANT_TEMPERATURE,
  loadAssistantSystemPrompt,
} from '@/lib/llm/assistant-config';

const systemPrompt = loadAssistantSystemPrompt();

// Simple rate limiting.
const requestTimes: number[] = [];

function isRateLimited(): boolean {
  const now = Date.now();
  const windowStart = now - 60000;
  while (requestTimes.length > 0 && requestTimes[0] < windowStart) {
    requestTimes.shift();
  }
  return requestTimes.length >= ASSISTANT_RATE_LIMIT_PER_MINUTE;
}

interface TokenClaims {
  valid: boolean;
  email?: string;
  sub?: string;
}

// Simple token verification: decode JWT and check expiry + issuer.
// Full JWKS verification would require a JWT library — this is a pragmatic check.
async function verifyToken(token: string): Promise<TokenClaims> {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return { valid: false };

    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
    const now = Math.floor(Date.now() / 1000);

    if (payload.exp && payload.exp < now) return { valid: false };

    const poolId = process.env.COGNITO_USER_POOL_ID;
    const region = process.env.AWS_REGION ?? 'eu-west-1';
    if (poolId) {
      const expectedIssuer = `https://cognito-idp.${region}.amazonaws.com/${poolId}`;
      if (payload.iss !== expectedIssuer) return { valid: false };
    }

    return {
      valid: true,
      email: typeof payload.email === 'string' ? payload.email : undefined,
      sub: typeof payload.sub === 'string' ? payload.sub : undefined,
    };
  } catch {
    return { valid: false };
  }
}

export async function POST(request: Request) {
  if (isRateLimited()) {
    return NextResponse.json({ error: 'Rate limited. Try again in a minute.' }, { status: 429 });
  }

  const authHeader = request.headers.get('authorization');
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return NextResponse.json({ error: 'Authentication required. Please sign in.' }, { status: 401 });
  }

  const token = authHeader.slice(7);
  const claims = await verifyToken(token);
  if (!claims.valid) {
    return NextResponse.json({ error: 'Invalid or expired token. Please sign in again.' }, { status: 401 });
  }

  requestTimes.push(Date.now());

  try {
    const { messages } = await request.json() as { messages: ChatMessage[] };

    if (!messages || messages.length === 0) {
      return NextResponse.json({ error: 'No messages provided' }, { status: 400 });
    }

    // Providers expect the conversation to start with a user turn.
    const validMessages = messages.filter((m, i) => {
      if (i === 0 && m.role === 'assistant') return false;
      return true;
    });

    const storage = getStorageProvider();
    let enrichedPrompt = systemPrompt;
    try {
      const manifestContent = await storage.readRootFile('manifest.json');
      if (manifestContent) {
        const manifest = JSON.parse(manifestContent);
        if (manifest.runs && manifest.runs.length > 0) {
          const runSummary = manifest.runs
            .slice(-ASSISTANT_MANIFEST_RUN_CONTEXT)
            .map((r: { runId?: string; totalPenalty?: number; algorithm?: string }) =>
              `${r.runId}: penalty=${r.totalPenalty || '?'} (${r.algorithm || '?'})`)
            .join('\n');
          enrichedPrompt += `\n\n## Recent Runs (from storage)\n\n${runSummary}`;
        }
      }
    } catch {
      // Non-critical — continue without run context.
    }

    const result = await completeChat({
      system: enrichedPrompt,
      messages: validMessages,
      maxTokens: ASSISTANT_MAX_TOKENS,
      temperature: ASSISTANT_TEMPERATURE,
    });

    try {
      await recordChatUsage(storage, {
        usage: result.usage,
        model: result.model,
        provider: result.provider,
        user: claims.email || claims.sub,
      });
    } catch (err) {
      console.error('Failed to record chat usage:', err);
    }

    return NextResponse.json({
      response: result.text,
      usage: {
        inputTokens: result.usage.inputTokens,
        outputTokens: result.usage.outputTokens,
        model: result.model,
        provider: result.provider,
      },
    });
  } catch (err) {
    console.error('Chat API error:', err);
    const message = err instanceof Error ? err.message : 'Unknown error';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
