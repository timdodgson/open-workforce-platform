import { NextResponse } from 'next/server';
import { readFileSync } from 'fs';
import { join } from 'path';
import { getStorageProvider } from '@/lib/storage';

const MODEL_ID = process.env.BEDROCK_MODEL_ID ?? 'anthropic.claude-3-haiku-20240307-v1:0';
const REGION = process.env.AWS_REGION ?? 'eu-west-1';

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

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export async function POST(request: Request) {
  if (isRateLimited()) {
    return NextResponse.json({ error: 'Rate limited. Try again in a minute.' }, { status: 429 });
  }
  requestTimes.push(Date.now());

  try {
    const { messages } = await request.json() as { messages: ChatMessage[] };

    if (!messages || messages.length === 0) {
      return NextResponse.json({ error: 'No messages provided' }, { status: 400 });
    }

    // Ensure conversation starts with a user message (Bedrock requirement).
    const validMessages = messages.filter((m, i) => {
      if (i === 0 && m.role === 'assistant') return false;
      return true;
    });

    // Optionally enrich system prompt with current run data.
    let enrichedPrompt = systemPrompt;
    try {
      const storage = getStorageProvider();
      const manifestContent = await storage.readRootFile('manifest.json');
      if (manifestContent) {
        const manifest = JSON.parse(manifestContent);
        if (manifest.runs && manifest.runs.length > 0) {
          const runSummary = manifest.runs
            .slice(-20) // Last 20 runs
            .map((r: any) => `${r.runId}: penalty=${r.totalPenalty || '?'} (${r.algorithm || '?'})`)
            .join('\n');
          enrichedPrompt += `\n\n## Recent Runs (from S3)\n\n${runSummary}`;
        }
      }
    } catch {
      // Non-critical — continue without run context.
    }

    // Dynamic require — only loads when the route is actually called.
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const { BedrockRuntimeClient, ConverseCommand } = require('@aws-sdk/client-bedrock-runtime');

    const client = new BedrockRuntimeClient({ region: REGION });

    const command = new ConverseCommand({
      modelId: MODEL_ID,
      system: [{ text: enrichedPrompt }],
      messages: validMessages.map(m => ({
        role: m.role,
        content: [{ text: m.content }],
      })),
      inferenceConfig: {
        maxTokens: 2048,
        temperature: 0.3,
      },
    });

    const response = await client.send(command);

    const responseText = response.output?.message?.content?.[0]?.text ?? 'No response from model.';

    return NextResponse.json({ response: responseText });
  } catch (err) {
    console.error('Chat API error:', err);
    const message = err instanceof Error ? err.message : 'Unknown error';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
