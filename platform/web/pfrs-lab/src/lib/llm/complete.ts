/**
 * LLM backends for the research assistant (/api/chat).
 * Default: Anthropic Messages API. Bedrock remains available when AWS enables it.
 */

export type LlmProviderName = 'anthropic' | 'bedrock';

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface LlmCompleteInput {
  system: string;
  messages: ChatMessage[];
  maxTokens?: number;
  temperature?: number;
}

function resolveProvider(): LlmProviderName {
  const raw = (process.env.LLM_PROVIDER || 'anthropic').toLowerCase().trim();
  if (raw === 'bedrock') return 'bedrock';
  return 'anthropic';
}

export function getLlmProviderName(): LlmProviderName {
  return resolveProvider();
}

async function completeAnthropic(input: LlmCompleteInput): Promise<string> {
  const apiKey = process.env.ANTHROPIC_API_KEY;
  if (!apiKey) {
    throw new Error(
      'ANTHROPIC_API_KEY is not set. Add it to .env.local (local) or SST/Lambda env (production).',
    );
  }

  const model =
    process.env.ANTHROPIC_MODEL_ID || 'claude-haiku-4-5-20251001';

  const res = await fetch('https://api.anthropic.com/v1/messages', {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      'x-api-key': apiKey,
      'anthropic-version': '2023-06-01',
    },
    body: JSON.stringify({
      model,
      max_tokens: input.maxTokens ?? 2048,
      temperature: input.temperature ?? 0.3,
      system: input.system,
      messages: input.messages.map((m) => ({
        role: m.role,
        content: m.content,
      })),
    }),
  });

  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const detail =
      typeof body?.error?.message === 'string'
        ? body.error.message
        : typeof body?.message === 'string'
          ? body.message
          : res.statusText;
    throw new Error(`Anthropic API ${res.status}: ${detail}`);
  }

  const text = body?.content?.find((b: { type?: string }) => b.type === 'text')?.text;
  if (typeof text === 'string' && text.length > 0) return text;
  return 'No response from model.';
}

async function completeBedrock(input: LlmCompleteInput): Promise<string> {
  const modelId =
    process.env.BEDROCK_MODEL_ID ?? 'eu.anthropic.claude-3-haiku-20240307-v1:0';
  const region = process.env.AWS_REGION ?? 'eu-west-1';

  // Dynamic require — only loads when Bedrock is selected.
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const { BedrockRuntimeClient, ConverseCommand } = require('@aws-sdk/client-bedrock-runtime');

  const client = new BedrockRuntimeClient({ region });
  const command = new ConverseCommand({
    modelId,
    system: [{ text: input.system }],
    messages: input.messages.map((m) => ({
      role: m.role,
      content: [{ text: m.content }],
    })),
    inferenceConfig: {
      maxTokens: input.maxTokens ?? 2048,
      temperature: input.temperature ?? 0.3,
    },
  });

  const response = await client.send(command);
  return response.output?.message?.content?.[0]?.text ?? 'No response from model.';
}

export async function completeChat(input: LlmCompleteInput): Promise<string> {
  const provider = resolveProvider();
  if (provider === 'bedrock') {
    return completeBedrock(input);
  }
  return completeAnthropic(input);
}
