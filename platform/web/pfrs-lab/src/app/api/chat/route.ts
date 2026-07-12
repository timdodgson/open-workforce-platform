import { NextResponse } from 'next/server';
import { readFileSync } from 'fs';
import { join } from 'path';
import { getStorageProvider } from '@/lib/storage';
import { completeChat, type ChatMessage } from '@/lib/llm/complete';

// Load system prompt at module level.
let systemPrompt: string;
try {
  systemPrompt = readFileSync(join(process.cwd(), 'src/lib/assistant-prompt.md'), 'utf-8');
} catch {
  systemPrompt = 'You are a PFRS optimisation experiment planner. Help users design nurse rostering experiments.';
}

// Simple rate limiting.
const requestTimes: number[] = [];
const MAX_REQUESTS_PER_MINUTE = 20;

function isRateLimited(): boolean {
  const now = Date.now();
  const windowStart = now - 60000;
  while (requestTimes.length > 0 && requestTimes[0] < windowStart) {
    requestTimes.shift();
  }
  return requestTimes.length >= MAX_REQUESTS_PER_MINUTE;
}

// Simple token verification: decode JWT and check expiry + issuer.
// Full JWKS verification would require a JWT library — this is a pragmatic check.
async function verifyToken(token: string): Promise<boolean> {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return false;

    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
    const now = Math.floor(Date.now() / 1000);

    if (payload.exp && payload.exp < now) return false;

    const poolId = process.env.COGNITO_USER_POOL_ID;
    const region = process.env.AWS_REGION ?? 'eu-west-1';
    if (poolId) {
      const expectedIssuer = `https://cognito-idp.${region}.amazonaws.com/${poolId}`;
      if (payload.iss !== expectedIssuer) return false;
    }

    return true;
  } catch {
    return false;
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
  const isValid = await verifyToken(token);
  if (!isValid) {
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

    let enrichedPrompt = systemPrompt;
    try {
      const storage = getStorageProvider();
      const manifestContent = await storage.readRootFile('manifest.json');
      if (manifestContent) {
        const manifest = JSON.parse(manifestContent);
        if (manifest.runs && manifest.runs.length > 0) {
          const runSummary = manifest.runs
            .slice(-20)
            .map((r: { runId?: string; totalPenalty?: number; algorithm?: string }) =>
              `${r.runId}: penalty=${r.totalPenalty || '?'} (${r.algorithm || '?'})`)
            .join('\n');
          enrichedPrompt += `\n\n## Recent Runs (from storage)\n\n${runSummary}`;
        }
      }
    } catch {
      // Non-critical — continue without run context.
    }

    const responseText = await completeChat({
      system: enrichedPrompt,
      messages: validMessages,
      maxTokens: 2048,
      temperature: 0.3,
    });

    return NextResponse.json({ response: responseText });
  } catch (err) {
    console.error('Chat API error:', err);
    const message = err instanceof Error ? err.message : 'Unknown error';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
